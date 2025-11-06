// Copyright 2024 Google, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: rrmcguinness (Ryan McGuinness)
//         jaycherian (Jay Cherian)

package workflow

import (
	goctx "context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"google.golang.org/api/iterator"
	"google.golang.org/genai"
)

type MediaEmbeddingGeneratorWorkflow struct {
	cor.BaseCommand
	genaiEmbedding         *genai.Models
	ModelName              string
	bigqueryClient         *bigquery.Client
	dataset                string
	mediaTable             string
	embeddingTable         string
	findEligibleMediaQuery string
}

func (m *MediaEmbeddingGeneratorWorkflow) StartTimer() {
	tracer := otel.Tracer("embedding-batch")
	ticker := time.NewTicker(60 * time.Second)
	closeTicker := make(chan struct{})

	// Create a timer to run embedding checks every 60 seconds
	go func(m *MediaEmbeddingGeneratorWorkflow) {
		for {
			select {
			case <-ticker.C:
				traceCtx, span := tracer.Start(goctx.Background(), "media-embeddings")
				chainCtx := cor.NewBaseContext()
				chainCtx.SetContext(traceCtx)
				m.Execute(chainCtx)
				if chainCtx.HasErrors() {
					span.SetStatus(codes.Error, "failed to execute embedding chain")
				} else {
					span.SetStatus(codes.Ok, "executed embeddings")
				}
				span.End()
			case <-closeTicker:
				ticker.Stop()
				return
			}
		}
	}(m)
}

func NewMediaEmbeddingGeneratorWorkflow(config *cloud.Config, serviceClients *cloud.ServiceClients) *MediaEmbeddingGeneratorWorkflow {

	fqMediaTableName := strings.Replace(serviceClients.BiqQueryClient.Dataset(config.BigQueryDataSource.DatasetName).Table(config.BigQueryDataSource.MediaTable).FullyQualifiedName(), ":", ".", -1)
	fqEmbeddingTable := strings.Replace(serviceClients.BiqQueryClient.Dataset(config.BigQueryDataSource.DatasetName).Table(config.BigQueryDataSource.EmbeddingTable).FullyQualifiedName(), ":", ".", -1)
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE ID NOT IN (SELECT MEDIA_ID FROM `%s`)", fqMediaTableName, fqEmbeddingTable)

	return &MediaEmbeddingGeneratorWorkflow{
		BaseCommand:            *cor.NewBaseCommand("media-embedding-generator"),
		genaiEmbedding:         serviceClients.EmbeddingModels["multi-lingual"],
		bigqueryClient:         serviceClients.BiqQueryClient,
		dataset:                config.BigQueryDataSource.DatasetName,
		mediaTable:             config.BigQueryDataSource.MediaTable,
		embeddingTable:         config.BigQueryDataSource.EmbeddingTable,
		findEligibleMediaQuery: query,
		ModelName:              config.EmbeddingModels["multi-lingual"].Model,
	}
}

func (m *MediaEmbeddingGeneratorWorkflow) IsExecutable(_ cor.Context) bool {
	return true
}

func (m *MediaEmbeddingGeneratorWorkflow) Execute(context cor.Context) {
	q := m.bigqueryClient.Query(m.findEligibleMediaQuery)
	it, err := q.Read(context.GetContext())
	if err != nil {
		context.AddError(m.GetName(), err)
		return
	}

	for {
		var value model.Media
		err := it.Next(&value)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			context.AddError(m.GetName(), err)
			return
		}

		toInsert := make([]*model.SegmentEmbedding, 0)

		for _, segment := range value.Segments {
			in := model.NewSegmentEmbedding(value.Id, segment.SequenceNumber, m.ModelName)
			contents := []*genai.Content{
				genai.NewContentFromText(segment.Script, genai.RoleUser),
			}

			resp, err := m.genaiEmbedding.EmbedContent(context.GetContext(), m.ModelName, contents, nil)
			if err != nil {
				context.AddError(m.GetName(), err)
				return
			}
			for _, f := range resp.Embeddings {
				for _, g := range f.Values {
					in.Embeddings = append(in.Embeddings, float64(g))
				}
			}
			toInsert = append(toInsert, in)
		}

		inserter := m.bigqueryClient.Dataset(m.dataset).Table(m.embeddingTable).Inserter()
		if err := inserter.Put(context.GetContext(), toInsert); err != nil {
			context.AddError(m.GetName(), err)
			return
		}
	}
}

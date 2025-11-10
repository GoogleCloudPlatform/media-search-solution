// Copyright 2025 Google, LLC
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
//         kingman (Charlie Wang)

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"google.golang.org/api/iterator"
	"google.golang.org/genai"
)

func generate_embeddings(genaiRunConfig *common.GenaiRunConfig) {
	stepConfig, err := common.NewGenaiStepConfig(common.EMBEDDING_STEP, genaiRunConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	stepConfig.StepLogic = generateEmbeddingsLogicFunc(stepConfig)
	stepConfig.RunStep()
}

func generateEmbeddingsLogicFunc(config *common.GenaiStepConfig) func() (string, error) {
	return func() (string, error) {
		mediaID := getMediaId(config)
		if mediaID == "" {
			return "", errors.New("could not retrieve mediaId from persist step")
		}
		bqDataset := config.GenaiRunConfig.CloudConfig.BigQueryDataSource.DatasetName
		bqMediaTable := config.GenaiRunConfig.CloudConfig.BigQueryDataSource.MediaTable
		bqEmbeddingTable := config.GenaiRunConfig.CloudConfig.BigQueryDataSource.EmbeddingTable
		bqClient := config.GenaiRunConfig.BigQueryClient

		// 1. Query BigQuery for the persisted Media object
		fqMediaTableName := strings.Replace(bqClient.Dataset(bqDataset).Table(bqMediaTable).FullyQualifiedName(), ":", ".", -1)

		q := bqClient.Query(fmt.Sprintf("SELECT * FROM `%s` WHERE id = '%s'",
			fqMediaTableName,
			mediaID))

		it, err := q.Read(config.BasicRunConfig.Ctx)
		if err != nil {
			return "", fmt.Errorf("error reading from BigQuery: %w", err)
		}

		var media model.Media
		err = it.Next(&media)
		if errors.Is(err, iterator.Done) {
			return "", fmt.Errorf("media with ID %s not found in BigQuery", mediaID)
		}
		if err != nil {
			return "", fmt.Errorf("error iterating BigQuery results: %w", err)
		}

		// 2. Generate embeddings for each segment
		numberOfSegments := len(media.Segments)
		toInsert := make([]*model.SegmentEmbedding, 0, numberOfSegments)
		embeddingModel := config.GenaiRunConfig.GenAIEmbedding
		modelName := config.GenaiRunConfig.CloudConfig.EmbeddingModels["multi-lingual"].Model
		for _, segment := range media.Segments {
			segmentEmbedding := model.NewSegmentEmbedding(media.Id, segment.SequenceNumber, modelName)
			contents := []*genai.Content{
				genai.NewContentFromText(segment.Script, genai.RoleUser),
			}

			resp, err := embeddingModel.EmbedContent(config.BasicRunConfig.Ctx, modelName, contents, nil)
			if err != nil {
				return "", fmt.Errorf("failed to generate embedding for segment %d: %w", segment.SequenceNumber, err)
			}

			for _, f := range resp.Embeddings {
				for _, g := range f.Values {
					segmentEmbedding.Embeddings = append(segmentEmbedding.Embeddings, float64(g))
				}
			}
			toInsert = append(toInsert, segmentEmbedding)
		}

		// 3. Insert embeddings into BigQuery
		inserter := bqClient.Dataset(bqDataset).Table(bqEmbeddingTable).Inserter()
		if numberOfSegments > 100 {
			for i := range numberOfSegments / 100 {
				start := i * 100
				end := min((i+1)*100, numberOfSegments)
				if err := inserter.Put(config.BasicRunConfig.Ctx, toInsert[start:end]); err != nil {
					return "", fmt.Errorf("failed to insert %dth batch ofembeddings into BigQuery: %w", i+1, err)
				}
			}

		} else {
			if err := inserter.Put(config.BasicRunConfig.Ctx, toInsert); err != nil {
				return "", fmt.Errorf("failed to insert embeddings into BigQuery: %w", err)
			}
		}

		return fmt.Sprintf("generated and persisted embeddings for %d segments", len(toInsert)), nil
	}
}

func getMediaId(config *common.GenaiStepConfig) string {
	inputParameter := []string{
		common.PERSIST_STEP,
	}
	inputValues := config.BasicRunConfig.GetStepsOutput(inputParameter)
	return inputValues[common.PERSIST_STEP]
}

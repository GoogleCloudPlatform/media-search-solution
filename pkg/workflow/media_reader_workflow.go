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
//         kingman (Charlie Wang)

package workflow

import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/commands"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"google.golang.org/genai"
)

type MediaReaderWorkflow struct {
	cor.BaseCommand
	config          *cloud.Config
	bigqueryClient  *bigquery.Client
	genaiClient     *genai.Client
	genaiModel      *cloud.QuotaAwareGenerativeAIModel
	storageClient   *storage.Client
	numberOfWorkers int
	templateService *cloud.TemplateService
	chain           cor.Chain
	ffprobeCommand  string
}

func (m *MediaReaderWorkflow) Execute(context cor.Context) {
	m.chain.Execute(context)
}

func (m *MediaReaderWorkflow) initializeChain() {
	const SummaryOutputParamName = "__summary_output__"
	const SegmentOutputParamName = "__segment_output__"
	const MediaOutputParamName = "__media_output__"
	const MediaLengthOutputParamName = "__media_length_output__"
	const ContentTypeOutputParamName = "__content_type_output__"

	out := cor.NewBaseChain(m.GetName())

	// Convert the Message to an Object
	out.AddCommand(commands.NewMediaTriggerToGCSObject("media-trigger-to-gcs-object"))

	// Get media length
	out.AddCommand(commands.NewMediaLengthCommand("get-media-length", m.ffprobeCommand, MediaLengthOutputParamName, m.config))

	// Determine the media content type
	out.AddCommand(commands.NewMediaContentTypeCommand("get-media-content-type", m.config, m.genaiModel, m.templateService, ContentTypeOutputParamName))

	// Generate Summary
	out.AddCommand(commands.NewMediaSummaryCreator("generate-media-summary", m.config, m.genaiModel, m.templateService, MediaLengthOutputParamName, ContentTypeOutputParamName))

	// Convert the JSON to a struct and save to the summaryOutputParam
	out.AddCommand(commands.NewMediaSummaryJsonToStruct("convert-media-summary", SummaryOutputParamName))

	// Create the segment extraction command
	segmentExtractor := commands.NewSegmentExtractor("extract-media-segments", m.genaiModel, m.templateService, m.numberOfWorkers, ContentTypeOutputParamName)
	segmentExtractor.BaseCommand.OutputParamName = SegmentOutputParamName
	out.AddCommand(segmentExtractor)

	// Assemble the output into a single media object
	out.AddCommand(commands.NewMediaAssembly("assemble-media-segments", SummaryOutputParamName, SegmentOutputParamName, MediaOutputParamName, MediaLengthOutputParamName))

	// Save media object to big query for async embedding job
	out.AddCommand(commands.NewMediaPersistToBigQuery(
		"write-to-bigquery",
		m.bigqueryClient,
		m.config.BigQueryDataSource.DatasetName,
		m.config.BigQueryDataSource.MediaTable, MediaOutputParamName))

	m.chain = out
}

func NewMediaReaderPipeline(
	config *cloud.Config,
	serviceClients *cloud.ServiceClients,
	agentModelName string,
	ffprobeCommand string,
	templateService *cloud.TemplateService) *MediaReaderWorkflow {

	pipeline := &MediaReaderWorkflow{
		BaseCommand:     *cor.NewBaseCommand("media-reader-pipeline"),
		config:          config,
		bigqueryClient:  serviceClients.BiqQueryClient,
		genaiClient:     serviceClients.GenAIClient,
		genaiModel:      serviceClients.AgentModels[agentModelName],
		storageClient:   serviceClients.StorageClient,
		numberOfWorkers: config.Application.ThreadPoolSize,
		templateService: templateService,
		ffprobeCommand:  ffprobeCommand,
	}
	pipeline.initializeChain()
	return pipeline
}

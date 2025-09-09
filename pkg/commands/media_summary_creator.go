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

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/genai"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
)

type MediaSummaryCreator struct {
	cor.BaseCommand
	config                     *cloud.Config
	generativeAIModel          *cloud.QuotaAwareGenerativeAIModel
	templateService            *cloud.TemplateService
	contentTypeParamName       string
	mediaLengthOutputParamName string
	geminiInputTokenCounter    metric.Int64Counter
	geminiOutputTokenCounter   metric.Int64Counter
	geminiRetryCounter         metric.Int64Counter
}

func NewMediaSummaryCreator(
	name string,
	config *cloud.Config,
	generativeAIModel *cloud.QuotaAwareGenerativeAIModel,
	templateService *cloud.TemplateService,
	mediaLengthOutputParamName string,
	contentTypeParamName string) *MediaSummaryCreator {

	out := &MediaSummaryCreator{
		BaseCommand:                *cor.NewBaseCommand(name),
		config:                     config,
		generativeAIModel:          generativeAIModel,
		templateService:            templateService,
		mediaLengthOutputParamName: mediaLengthOutputParamName,
		contentTypeParamName:       contentTypeParamName,
	}

	out.geminiInputTokenCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.input", out.GetName()))
	out.geminiOutputTokenCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.ouput", out.GetName()))
	out.geminiRetryCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.retry", out.GetName()))

	return out
}

func (t *MediaSummaryCreator) GenerateParams(context cor.Context) map[string]interface{} {
	mediaLengthInSeconds := context.Get(t.mediaLengthOutputParamName).(int)
	params := make(map[string]interface{})

	// Create a string representation of the categories
	catStr := ""
	for key, cat := range t.config.Categories {
		catStr += key + " - " + cat.Definition + "; "
	}
	params["CATEGORIES"] = t.config.Categories

	exampleSummary, _ := json.Marshal(model.GetExampleSummary())
	params["EXAMPLE_JSON"] = string(exampleSummary)
	params["VIDEO_LENGTH"] = fmt.Sprintf("%d", mediaLengthInSeconds)
	return params
}

func (t *MediaSummaryCreator) Execute(context cor.Context) {
	gcsFile := context.Get(cloud.GetGCSObjectName()).(*cloud.GCSObject)
	gcsFileLink := fmt.Sprintf("gs://%s/%s", gcsFile.Bucket, gcsFile.Name)
	mediaType := context.Get(t.contentTypeParamName).(string)

	var buffer bytes.Buffer
	err := t.templateService.GetTemplateBy(mediaType).SummaryPrompt.Execute(&buffer, t.GenerateParams(context))
	if err != nil {
		t.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(t.GetName(), err)
		return
	}

	contents := []*genai.Content{
		{Parts: []*genai.Part{
			genai.NewPartFromText(buffer.String()),
			genai.NewPartFromURI(gcsFileLink, gcsFile.MIMEType),
		},
			Role: "user"},
	}

	// Get the response
	out, err := cloud.GenerateMultiModalResponse(context.GetContext(), t.geminiInputTokenCounter, t.geminiOutputTokenCounter, t.geminiRetryCounter, 0, t.generativeAIModel, t.templateService.GetTemplateBy(mediaType).SystemInstructions, contents, model.NewMediaSummarySchema())
	if err != nil {
		t.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(t.GetName(), err)
		return
	}
	t.GetSuccessCounter().Add(context.GetContext(), 1)
	context.Add(t.GetOutputParam(), out)
}

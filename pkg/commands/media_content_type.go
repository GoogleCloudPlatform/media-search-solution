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
// Author: kingman (Charlie Wang)

package commands

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/genai"
)

type MediaContentTypeCommand struct {
	cor.BaseCommand
	templateService          *cloud.TemplateService
	config                   *cloud.Config
	generativeAIModel        *cloud.QuotaAwareGenerativeAIModel
	geminiInputTokenCounter  metric.Int64Counter
	geminiOutputTokenCounter metric.Int64Counter
	geminiRetryCounter       metric.Int64Counter
}

func NewMediaContentTypeCommand(
	name string,
	config *cloud.Config,
	generativeAIModel *cloud.QuotaAwareGenerativeAIModel,
	templateService *cloud.TemplateService,
	outputParamName string) *MediaContentTypeCommand {

	out := MediaContentTypeCommand{
		BaseCommand:       *cor.NewBaseCommand(name),
		config:            config,
		generativeAIModel: generativeAIModel,
		templateService:   templateService,
	}

	out.geminiInputTokenCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.input", out.GetName()))
	out.geminiOutputTokenCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.ouput", out.GetName()))
	out.geminiRetryCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.retry", out.GetName()))
	out.OutputParamName = outputParamName

	return &out
}

func (c *MediaContentTypeCommand) Execute(context cor.Context) {
	gcsFile := context.Get(cloud.GetGCSObjectName()).(*cloud.GCSObject)
	gcsFileLink := fmt.Sprintf("gs://%s/%s", gcsFile.Bucket, gcsFile.Name)

	params := make(map[string]interface{})

	params["CONTENT_TYPES"] = c.config.ContentType.Types

	var buffer bytes.Buffer
	err := c.templateService.GetContentTypeTemplate().Execute(&buffer, params)
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), err)
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
	out, err := cloud.GenerateMultiModalResponse(context.GetContext(), c.geminiInputTokenCounter, c.geminiOutputTokenCounter, c.geminiRetryCounter, 0, c.generativeAIModel, "", contents, nil)
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), err)
		return
	}

	out = strings.TrimSpace(out)

	valid := false
	for _, value := range c.config.ContentType.Types {
		if strings.Contains(strings.ToLower(out), strings.ToLower(value)) {
			out = value
			valid = true
			break
		}
	}
	if !valid {
		log.Printf("LLM returned an invalid content type '%s', defaulting to '%s'", out, c.config.ContentType.DefaultType)
		out = c.config.ContentType.DefaultType
	}
	c.GetSuccessCounter().Add(context.GetContext(), 1)
	context.Add(c.GetOutputParam(), out)
	context.Add(cor.CtxOut, out)
}

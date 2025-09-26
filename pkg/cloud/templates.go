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
// Author: kingman (Charlie Wang)

package cloud

import "text/template"

type TemplateService struct {
	config              *Config
	templateByMediaType map[string]*PromptTemplate
	contentTypeTemplate *template.Template
}

func NewTemplateService(config *Config) *TemplateService {
	out := &TemplateService{
		config: config,
	}
	out.UpdateTemplates()
	return out
}

func (t *TemplateService) GetTemplateBy(mediaType string) *PromptTemplate {
	return t.templateByMediaType[mediaType]
}

func (t *TemplateService) GetContentTypeTemplate() *template.Template {
	return t.contentTypeTemplate
}

func (t *TemplateService) UpdateTemplates() {
	t.templateByMediaType = GetTemplateByMediaType(t.config)
	t.contentTypeTemplate = GetContentTypeTemplate(t.config)
}

func GetTemplateByMediaType(config *Config) map[string]*PromptTemplate {
	templateByMediaType := make(map[string]*PromptTemplate)
	for mediaType := range config.PromptTemplates {
		systemInstruction := config.PromptTemplates[mediaType].SystemInstructions
		summaryTemplate, err := template.New("summary-template").Parse(config.PromptTemplates[mediaType].SummaryPrompt)
		if err != nil {
			panic(err)
		}
		segmentTemplate, err := template.New("segment-template").Parse(config.PromptTemplates[mediaType].SegmentPrompt)
		if err != nil {
			panic(err)
		}
		templateByMediaType[mediaType] = &PromptTemplate{
			SystemInstructions: systemInstruction,
			SummaryPrompt:      summaryTemplate,
			SegmentPrompt:      segmentTemplate,
		}
	}
	return templateByMediaType
}

func GetContentTypeTemplate(config *Config) *template.Template {
	contentTypeTemplate, err := template.New("content-type-template").Parse(config.ContentType.PromptTemplate)
	if err != nil {
		panic(err)
	}
	return contentTypeTemplate
}

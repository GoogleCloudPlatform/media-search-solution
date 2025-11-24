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
	"bytes"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
)

const (
	CONTENT_TYPE_STEP_MODEL            = "creative-flash"
	CONTENT_TYPE_ANALYSIS_START_OFFSET = 0
	CONTENT_TYPE_ANALYSIS_END_OFFSET   = 30
)

func get_content_type(genaiRunConfig *common.GenaiRunConfig) {
	stepConfig, err := common.NewGenaiStepConfig(common.CONTENT_TYPE_STEP, genaiRunConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	stepConfig.StepLogic = getContentTypLogicFunc(stepConfig)
	stepConfig.RunStep()
}

func getContentTypLogicFunc(config *common.GenaiStepConfig) func() (string, error) {
	return func() (string, error) {

		params := make(map[string]interface{})
		params["CONTENT_TYPES"] = strings.Join(config.GenaiRunConfig.CloudConfig.ContentType.Types, "\n")

		var buffer bytes.Buffer
		if err := config.GenaiRunConfig.TemplateService.GetContentTypeTemplate().Execute(&buffer, params); err != nil {
			return "", err
		}

		generateContentConfig := &common.GenerateContentConfig{
			ModelName:         CONTENT_TYPE_STEP_MODEL,
			SystemInstruction: config.GenaiRunConfig.AgentModels[CONTENT_TYPE_STEP_MODEL].GenerativeContentConfig.SystemInstruction.Parts[0].Text,
			Prompt:            buffer.String(),
			Schema:            nil,
			StartOffset:       CONTENT_TYPE_ANALYSIS_START_OFFSET,
			EndOffset:         CONTENT_TYPE_ANALYSIS_END_OFFSET,
		}

		out, err := config.GenerateContentWithClippingInterval(generateContentConfig)

		if err != nil {
			log.Fatal(err)
		}

		out = strings.TrimSpace(out)

		valid := false
		for _, value := range config.GenaiRunConfig.CloudConfig.ContentType.Types {
			if strings.Contains(strings.ToLower(out), strings.ToLower(value)) {
				out = value
				valid = true
				break
			}
		}
		if !valid {
			log.Printf("LLM returned an invalid content type '%s', defaulting to '%s'", out, config.GenaiRunConfig.CloudConfig.ContentType.DefaultType)
			out = config.GenaiRunConfig.CloudConfig.ContentType.DefaultType
		}
		return out, nil

	}
}

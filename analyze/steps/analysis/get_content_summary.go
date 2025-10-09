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
	"encoding/json"
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"google.golang.org/genai"
)

const (
	CONTENT_SUMMARY_STEP_MODEL = "creative-flash"
)

func get_content_summary(genaiRunConfig *common.GenaiRunConfig) {
	stepConfig, err := common.NewGenaiStepConfig(common.CONTENT_SUMMARY_STEP, genaiRunConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	stepConfig.StepLogic = getContentSummaryLogicFunc(stepConfig)
	stepConfig.RunStep()
}

func getContentSummaryLogicFunc(config *common.GenaiStepConfig) func() (string, error) {
	return func() (string, error) {
		// Ensure that the required previous steps are completed
		inputParameter := []string{
			common.CONTENT_LENGTH_STEP,
			common.CONTENT_TYPE_STEP,
		}

		inputValues := config.BasicRunConfig.GetStepsOutput(inputParameter)

		for _, stepKey := range inputParameter {
			if _, ok := inputValues[stepKey]; !ok {
				return "", fmt.Errorf("missing required input from step: %s", stepKey)
			}
		}

		prompt, err := generatePrompt(config, inputValues)
		if err != nil {
			return "", err
		}
		contents := []*genai.Content{
			{Parts: []*genai.Part{
				genai.NewPartFromText(prompt),
			},
				Role: "user"},
		}
		systemInstructions := genai.NewContentFromText(config.GenaiRunConfig.TemplateService.GetTemplateBy(inputValues[common.CONTENT_TYPE_STEP]).SystemInstructions, genai.RoleUser)
		genaiContentCache, err := config.GenaiRunConfig.GetGenaiContentCache(
			CONTENT_SUMMARY_STEP_MODEL,
			inputValues[common.CONTENT_TYPE_STEP],
			systemInstructions,
		)
		if err != nil {
			return "", err
		}
		out, err := cloud.GenerateMultiModalResponse(
			config.BasicRunConfig.Ctx,
			config.Counters.InputCounter,
			config.Counters.OutputCounter,
			config.Counters.RetryCounter, 0,
			config.GenaiRunConfig.AgentModels[CONTENT_SUMMARY_STEP_MODEL],
			"",
			genaiContentCache.Name,
			contents,
			model.NewMediaSummarySchema(),
		)
		if err != nil {
			return "", err
		}

		out, err = normalizeOutPut(config, out)
		if err != nil {
			return "", err
		}
		return out, nil
	}
}

func normalizeOutPut(config *common.GenaiStepConfig, rawOutput string) (string, error) {
	obj := &model.MediaSummary{}
	err := json.Unmarshal([]byte(rawOutput), &obj)
	if err != nil {
		return "", err
	}
	obj.MediaUrl = fmt.Sprintf("https://storage.mtls.cloud.google.com/%s/%s", config.BasicRunConfig.InputBucket, config.BasicRunConfig.InputFile)
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(objBytes), nil
}

func generatePrompt(config *common.GenaiStepConfig, inputValues map[string]string) (string, error) {
	templateParams := make(map[string]interface{})

	catStr := ""
	for key, cat := range config.GenaiRunConfig.CloudConfig.Categories {
		catStr += key + " - " + cat.Definition + "; "
	}

	exampleSummary, err := json.Marshal(model.GetExampleSummary())
	if err != nil {
		return "", err
	}

	templateParams["CATEGORIES"] = catStr
	templateParams["EXAMPLE_JSON"] = string(exampleSummary)
	templateParams["VIDEO_LENGTH"] = inputValues[common.CONTENT_LENGTH_STEP]

	var buffer bytes.Buffer
	if err := config.GenaiRunConfig.TemplateService.GetTemplateBy(inputValues[common.CONTENT_TYPE_STEP]).SummaryPrompt.Execute(&buffer, templateParams); err != nil {
		return "", err
	}

	return buffer.String(), nil

}

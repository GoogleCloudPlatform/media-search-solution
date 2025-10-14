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
	"math"
	"strconv"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"google.golang.org/genai"
)

const (
	CONTENT_SUMMARY_STEP_MODEL = "creative-flash"
	maxRetries                 = 3
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

		var stepErr error
		for i := range maxRetries {
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
				stepErr = err
				continue
			}
			normalizedOutput, err := normalizeAndValidateOutput(config, out, inputValues[common.CONTENT_LENGTH_STEP])
			if err == nil {
				return normalizedOutput, nil
			}
			stepErr = err
			log.Printf("Content summary validation failed on attempt %d: %v", i+1, stepErr)
		}
		return "", fmt.Errorf("content summary generation and validation failed after %d attempts: %w", maxRetries, stepErr)
	}
}

func normalizeAndValidateOutput(config *common.GenaiStepConfig, rawOutput string, videoLengthStr string) (string, error) {
	obj := &model.MediaSummary{}
	if err := json.Unmarshal([]byte(rawOutput), obj); err != nil {
		return "", fmt.Errorf("failed to unmarshal content summary: %w", err)
	}

	videoLength, _ := strconv.Atoi(videoLengthStr)

	if err := validateContentSummary(obj, videoLength); err != nil {
		return "", err
	}

	obj.MediaUrl = fmt.Sprintf("https://storage.mtls.cloud.google.com/%s/%s", config.BasicRunConfig.InputBucket, config.BasicRunConfig.InputFile)
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal validated content summary: %w", err)
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
	templateParams["VIDEO_END_TIMESTAMP"] = convertSecondsToHHMMSS(inputValues[common.CONTENT_LENGTH_STEP])

	var buffer bytes.Buffer
	if err := config.GenaiRunConfig.TemplateService.GetTemplateBy(inputValues[common.CONTENT_TYPE_STEP]).SummaryPrompt.Execute(&buffer, templateParams); err != nil {
		return "", err
	}

	return buffer.String(), nil

}

func convertSecondsToHHMMSS(secondsStr string) string {
	s, err := strconv.Atoi(secondsStr)
	if err != nil {
		return ""
	}
	hours := s / 3600
	minutes := (s % 3600) / 60
	seconds := s % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func timeToSeconds(ts string) (int, error) {
	var h, m, s int
	_, err := fmt.Sscanf(ts, "%d:%d:%d", &h, &m, &s)
	if err != nil {
		return 0, fmt.Errorf("invalid time format '%s': %w", ts, err)
	}
	return h*3600 + m*60 + s, nil
}

func validateContentSummary(summary *model.MediaSummary, videoLength int) error {
	if len(summary.SegmentTimeStamps) == 0 {
		return fmt.Errorf("no segment timestamps found")
	}

	// 1. Ensure the first segment starts at 0 seconds
	firstStart, err := timeToSeconds(summary.SegmentTimeStamps[0].Start)
	if err != nil {
		return fmt.Errorf("invalid start time for first segment: %w", err)
	}
	if firstStart > 1 {
		return fmt.Errorf("first segment does not start at 0 seconds (with 1s tolerance), but at %d", firstStart)
	}

	// 1. Ensure the last segment ends at the length of the video
	endTimestamp, err := timeToSeconds(summary.SegmentTimeStamps[len(summary.SegmentTimeStamps)-1].End)
	if err != nil {
		return fmt.Errorf("invalid end time for last segment: %w", err)
	}
	if math.Abs(float64(endTimestamp-videoLength)) > 1 {
		return fmt.Errorf("last segment does not end at the length of the video %ds (with 1s tolerance), but at %ds", videoLength, endTimestamp)
	}

	var prevEnd int = 0
	for i, segment := range summary.SegmentTimeStamps {
		start, err := timeToSeconds(segment.Start)
		if err != nil {
			return fmt.Errorf("segment %d: invalid start time '%s': %w", i+1, segment.Start, err)
		}
		end, err := timeToSeconds(segment.End)
		if err != nil {
			return fmt.Errorf("segment %d: invalid end time '%s': %w", i+1, segment.End, err)
		}

		// 3. for each of the segment ensure the start is before the end timestamp
		if start >= end {
			return fmt.Errorf("segment %d: start time %s is not before end time %s", i+1, segment.Start, segment.End)
		}

		// 4. the start of one segement is same as the end the previous segment, here we can tolerate 1 second diff
		if i > 0 && start-prevEnd > 1 {
			return fmt.Errorf("segment %d: gap detected, start time %s does not follow previous end time %d within 1s tolerance", i+1, segment.Start, prevEnd)
		}
		prevEnd = end
	}

	return nil
}

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
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"google.golang.org/genai"
)

const (
	SEGMENT_SUMMARY_STEP_MODEL = "creative-flash"
)

func get_segment_summary(genaiRunConfig *common.GenaiRunConfig, mediaSummary *model.MediaSummary, contentType string, segmentSequenceNumber int) (string, error) {
	stepConfig, err := common.NewGenaiStepConfig(getSegmentSummaryStepKey(segmentSequenceNumber), genaiRunConfig, nil)
	if err != nil {
		return "", err
	}

	stepConfig.StepLogic = getSegmentSummaryLogicFunc(stepConfig, mediaSummary, contentType, segmentSequenceNumber)
	return stepConfig.StepLogic()
}

func getSegmentSummaryLogicFunc(config *common.GenaiStepConfig, mediaSummary *model.MediaSummary, contentType string, segmentSequenceNumber int) func() (string, error) {

	return func() (string, error) {
		prompt, err := generateSegmentSummaryPrompt(config, mediaSummary, contentType, segmentSequenceNumber)
		if err != nil {
			return "", err
		}

		segmentTimeSpan := mediaSummary.SegmentTimeStamps[segmentSequenceNumber]

		startOffset, err := getTimeDurationFrom(segmentTimeSpan.Start)
		if err != nil {
			return "", fmt.Errorf("invalid start timestamp format for segment %d: %w", segmentSequenceNumber, err)
		}
		endOffset, err := getTimeDurationFrom(segmentTimeSpan.End)
		if err != nil {
			return "", fmt.Errorf("invalid end timestamp format for segment %d: %w", segmentSequenceNumber, err)
		}

		genaiContentCacheName, err := getSegmentSummaryContentCacheName(config, contentType, startOffset, endOffset)
		if err != nil {
			return "", err
		}

		contents := []*genai.Content{
			{Parts: []*genai.Part{
				genai.NewPartFromText(prompt),
			},
				Role: "user"},
		}

		out, err := cloud.GenerateMultiModalResponse(
			config.BasicRunConfig.Ctx,
			config.Counters.InputCounter,
			config.Counters.OutputCounter,
			config.Counters.RetryCounter, 0,
			config.GenaiRunConfig.AgentModels[SEGMENT_SUMMARY_STEP_MODEL],
			"",
			genaiContentCacheName,
			contents,
			model.NewSegmentExtractorSchema())
		if err != nil {
			return "", err
		}
		return out, nil
	}
}

func getSegmentSummaryContentCacheName(config *common.GenaiStepConfig, contentType string, startOffset int, endOffset int) (string, error) {
	systemInstructions := genai.NewContentFromText(config.GenaiRunConfig.TemplateService.GetTemplateBy(contentType).SystemInstructions, genai.RoleUser)
	genaiContentCache, err := config.GetGenaiContentCacheWithChunk(
		SEGMENT_SUMMARY_STEP_MODEL,
		systemInstructions,
		startOffset,
		endOffset,
	)
	if err != nil {
		return "", err
	}
	return genaiContentCache.Name, nil
}

func getTimeDurationFrom(timestamp string) (int, error) {
	parts := strings.Split(timestamp, ":")
	var hours, minutes, seconds int
	var err error

	switch len(parts) {
	case 2: // "mm:ss"
		minutes, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %w", err)
		}
		seconds, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %w", err)
		}
	case 3: // "hh:mm:ss"
		hours, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %w", err)
		}
		minutes, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %w", err)
		}
		seconds, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %w", err)
		}
	default:
		return 0, fmt.Errorf("invalid time format: %s", timestamp)
	}
	return hours*3600 + minutes*60 + seconds, nil

}

// getSegmentSummaryStepKey generates the step key for a segment summary.
// The segmentSequenceNumber is the 0-based index of the segment in the array,
// but for storage and display purposes, it is converted to a 1-based index.
func getSegmentSummaryStepKey(segmentSequenceNumber int) string {
	return common.SEGMENT_SUMMARY_STEP_PREFIX + strconv.Itoa(segmentSequenceNumber+1)
}

func generateSegmentSummaryPrompt(config *common.GenaiStepConfig, mediaSummary *model.MediaSummary, contentType string, segmentSequanceNumber int) (string, error) {
	template := config.GenaiRunConfig.TemplateService.GetTemplateBy(contentType).SegmentPrompt
	templateParams := make(map[string]string)
	exampleSegment := model.GetExampleSegment()
	exampleJson, _ := json.Marshal(exampleSegment)
	exampleText := string(exampleJson)
	castString := ""
	for _, cast := range mediaSummary.Cast {
		castString += fmt.Sprintf("%s - %s\n", cast.CharacterName, cast.ActorName)
	}
	timeSpan := mediaSummary.SegmentTimeStamps[segmentSequanceNumber]
	summaryText := fmt.Sprintf("Title:%s\nSummary:\n\n%s\nCast:\n\n%v\n", mediaSummary.Title, mediaSummary.Summary, castString)
	templateParams["SEQUENCE"] = fmt.Sprintf("%d", segmentSequanceNumber+1)
	templateParams["SUMMARY_DOCUMENT"] = summaryText
	templateParams["TIME_START"] = timeSpan.Start
	templateParams["TIME_END"] = timeSpan.End
	templateParams["EXAMPLE_JSON"] = exampleText

	var doc bytes.Buffer
	if err := template.Execute(&doc, templateParams); err != nil {
		return "", err
	}
	return doc.String(), nil
}

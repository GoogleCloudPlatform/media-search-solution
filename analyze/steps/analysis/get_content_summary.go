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
	maxRetries                 = 5
	CHUNK_LENGTH_SEC           = 600
)

type ContentSummaryConfigAPI interface {
	getStepKey() string
	getLength() int
	getLengthStr() string
	getContentType() string
	isChunk() bool
	getStartOffsetSec() int
	getEndOffsetSec() int
}

type ContentSummaryConfig struct {
	ContentLength int
	ContentType   string
}

func (c ContentSummaryConfig) getStepKey() string {
	return common.CONTENT_SUMMARY_STEP
}

func (c ContentSummaryConfig) getLength() int {
	return c.ContentLength
}

func (c ContentSummaryConfig) getLengthStr() string {
	return strconv.Itoa(c.ContentLength)
}

func (c ContentSummaryConfig) getContentType() string {
	return c.ContentType
}

func (c ContentSummaryConfig) isChunk() bool {
	return false
}

func (c ContentSummaryConfig) getStartOffsetSec() int {
	return 0
}

func (c ContentSummaryConfig) getEndOffsetSec() int {
	return c.ContentLength
}

type ChunkConfig struct {
	ContentSummaryConfig
	StartOffsetSec int
	EndOffsetSec   int
	ChunkIndex     int
}

func (c ChunkConfig) getStartOffsetSec() int {
	return c.StartOffsetSec
}

func (c ChunkConfig) getEndOffsetSec() int {
	return c.EndOffsetSec
}

func (c ChunkConfig) getLength() int {
	return c.EndOffsetSec - c.StartOffsetSec
}

func (c ChunkConfig) getLengthStr() string {
	return strconv.Itoa(c.EndOffsetSec - c.StartOffsetSec)
}

func (c ChunkConfig) isChunk() bool {
	return true
}

func (c ChunkConfig) getStepKey() string {
	return fmt.Sprintf("%s_%d_%d", common.CONTENT_SUMMARY_STEP, c.StartOffsetSec, c.EndOffsetSec)
}

func (c ChunkConfig) convertChunkTimeStampToTimeStamp(chunkTimeStamp string) string {
	chunkTimeStampSec, err := timeToSeconds(chunkTimeStamp)
	if err != nil {
		return ""
	}
	return convertIntSecondsToHHMMSS(chunkTimeStampSec + c.StartOffsetSec)

}

func get_content_summary(genaiRunConfig *common.GenaiRunConfig) {
	contentSummaryConfig, err := getContentSummaryConfig(genaiRunConfig)
	if err != nil {
		log.Fatal(err)
	}
	if contentSummaryConfig.ContentLength > CHUNK_LENGTH_SEC && contentSummaryConfig.ContentLength-CHUNK_LENGTH_SEC > 60 {
		chunkConfigs := make([]*ChunkConfig, 0)
		numberOfChunks := contentSummaryConfig.ContentLength / CHUNK_LENGTH_SEC
		remainingSeconds := contentSummaryConfig.ContentLength % CHUNK_LENGTH_SEC
		// If remaining seconds is more than 60s, we create an additional chunk
		if remainingSeconds > 60 {
			numberOfChunks += 1
		}
		for chunkIndex := range numberOfChunks {
			endSecOffSet := (chunkIndex + 1) * CHUNK_LENGTH_SEC
			// For the last chunk, we set the end offset to content length if the remaining seconds is less than 60s
			if endSecOffSet > contentSummaryConfig.ContentLength || contentSummaryConfig.ContentLength-endSecOffSet <= 60 {
				endSecOffSet = contentSummaryConfig.ContentLength
			}
			chunkConfig := &ChunkConfig{
				ContentSummaryConfig: *contentSummaryConfig,
				StartOffsetSec:       chunkIndex * CHUNK_LENGTH_SEC,
				EndOffsetSec:         endSecOffSet,
				ChunkIndex:           chunkIndex,
			}
			chunkConfigs = append(chunkConfigs, chunkConfig)
			get_content_summary_dynamic(genaiRunConfig, chunkConfig)
		}
		consolidate_chunk_summaries(genaiRunConfig, chunkConfigs, contentSummaryConfig)

	} else {
		get_content_summary_dynamic(genaiRunConfig, contentSummaryConfig)
	}
}

func consolidate_chunk_summaries(genaiRunConfig *common.GenaiRunConfig, chunkConfigs []*ChunkConfig, summaryConfig ContentSummaryConfigAPI) {
	stepConfig, err := common.NewGenaiStepConfig(summaryConfig.getStepKey(), genaiRunConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	stepConfig.StepLogic = func() (string, error) {
		return consolidateChunkSummaries(genaiRunConfig, chunkConfigs, summaryConfig)
	}

	stepConfig.RunStep()
}

func get_content_summary_dynamic(genaiRunConfig *common.GenaiRunConfig, summaryConfig ContentSummaryConfigAPI) {
	stepConfig, err := common.NewGenaiStepConfig(summaryConfig.getStepKey(), genaiRunConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	stepConfig.StepLogic = getContentSummaryLogicFunc(stepConfig, summaryConfig)
	stepConfig.RunStep()
}

func getContentSummaryConfig(genaiRunConfig *common.GenaiRunConfig) (*ContentSummaryConfig, error) {
	inputParameter := []string{
		common.CONTENT_LENGTH_STEP,
		common.CONTENT_TYPE_STEP,
	}
	inputValues := genaiRunConfig.BasicRunConfig.GetStepsOutput(inputParameter)
	for _, stepKey := range inputParameter {
		if _, ok := inputValues[stepKey]; !ok {
			return nil, fmt.Errorf("missing required input from step: %s", stepKey)
		}
	}
	videoLength := inputValues[common.CONTENT_LENGTH_STEP]
	videoLengthSec, err := strconv.Atoi(videoLength)
	if err != nil {
		return nil, fmt.Errorf("invalid content length: %s", videoLength)
	}

	contentType := inputValues[common.CONTENT_TYPE_STEP]
	return &ContentSummaryConfig{
		ContentLength: videoLengthSec,
		ContentType:   contentType,
	}, nil
}

func getContentSummaryLogicFunc(config *common.GenaiStepConfig, summaryConfig ContentSummaryConfigAPI) func() (string, error) {
	return func() (string, error) {
		normalizedOutput, err := doSummaryGeneration(config, summaryConfig)
		if err != nil {
			return "", err
		}
		return normalizedOutput, nil
	}
}

func consolidateChunkSummaries(config *common.GenaiRunConfig, chunkConfigs []*ChunkConfig, summaryConfig ContentSummaryConfigAPI) (string, error) {
	chunkStepKeys := make([]string, len(chunkConfigs))

	for i, chunkConfig := range chunkConfigs {
		chunkStepKeys[i] = chunkConfig.getStepKey()
	}

	summaryPersistStrs := config.BasicRunConfig.GetStepsOutput(chunkStepKeys)
	summaryObjs := make([]*model.MediaSummary, len(summaryPersistStrs))

	consolidatedSegments := make([]*model.TimeSpan, 0)

	for i, chunkConfig := range chunkConfigs {
		summaryStr := summaryPersistStrs[chunkConfig.getStepKey()]
		summaryObj := &model.MediaSummary{}
		json.Unmarshal([]byte(summaryStr), summaryObj)
		for _, segment := range summaryObj.SegmentTimeStamps {
			startTime := chunkConfig.convertChunkTimeStampToTimeStamp(segment.Start)
			endTime := chunkConfig.convertChunkTimeStampToTimeStamp(segment.End)
			if startTime == "" || endTime == "" {
				continue
			}
			segment.Start = startTime
			segment.End = endTime
			consolidatedSegments = append(consolidatedSegments, segment)
		}
		summaryObjs[i] = summaryObj
	}

	consolidatedSummary := &model.MediaSummary{}
	consolidatedSummary.Title = summaryObjs[0].Title
	consolidatedSummary.Category = summaryObjs[0].Category
	consolidatedSummary.Summary = summaryObjs[0].Summary
	consolidatedSummary.LengthInSeconds = summaryConfig.getLength()
	consolidatedSummary.Director = summaryObjs[0].Director
	consolidatedSummary.ReleaseYear = summaryObjs[0].ReleaseYear
	consolidatedSummary.Genre = summaryObjs[0].Genre
	consolidatedSummary.Rating = summaryObjs[0].Rating
	consolidatedSummary.MediaUrl = summaryObjs[0].MediaUrl

	casts := make([]*model.CastMember, 0)
	for _, summaryObj := range summaryObjs {
		casts = append(casts, summaryObj.Cast...)
	}
	consolidatedSummary.Cast = casts

	consolidatedSummary.SegmentTimeStamps = consolidatedSegments
	objBytes, err := json.Marshal(consolidatedSummary)
	if err != nil {
		return "", fmt.Errorf("failed to marshal consolidated content summary: %w", err)
	}
	return string(objBytes), nil
}

func doSummaryGeneration(config *common.GenaiStepConfig, summaryConfig ContentSummaryConfigAPI) (string, error) {
	contentType := summaryConfig.getContentType()

	prompt, err := generatePrompt(config, summaryConfig.getLengthStr(), contentType)
	if err != nil {
		return "", err
	}

	videoLength := summaryConfig.getLengthStr()

	systemInstructions := genai.NewContentFromText(config.GenaiRunConfig.TemplateService.GetTemplateBy(contentType).SystemInstructions, genai.RoleUser)
	contents := []*genai.Content{
		{Parts: []*genai.Part{
			genai.NewPartFromText(prompt),
		},
			Role: genai.RoleUser},
	}
	var genaiContentCache *genai.CachedContent
	if summaryConfig.isChunk() {
		genaiContentCache, err = config.GetGenaiContentCacheWithChunk(
			CONTENT_SUMMARY_STEP_MODEL,
			systemInstructions,
			summaryConfig.getStartOffsetSec(),
			summaryConfig.getEndOffsetSec(),
		)
		if err != nil {
			return "", err
		}
	} else {
		genaiContentCache, err = config.GetGenaiContentCache(
			CONTENT_SUMMARY_STEP_MODEL,
			systemInstructions,
		)
		if err != nil {
			return "", err
		}
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

		normalizedOutput, err := normalizeAndValidateOutput(config, out, videoLength)
		if err == nil {
			return normalizedOutput, nil
		}
		stepErr = err
		log.Printf("Content summary validation failed on attempt %d: %v", i+1, stepErr)
	}
	return "", fmt.Errorf("content summary generation and validation failed after %d attempts: %w", maxRetries, stepErr)
}

func normalizeAndValidateOutput(config *common.GenaiStepConfig, rawOutput string, videoLengthStr string) (string, error) {
	obj := &model.MediaSummary{}
	if err := json.Unmarshal([]byte(rawOutput), obj); err != nil {
		return "", fmt.Errorf("failed to unmarshal content summary: %w", err)
	}

	videoLength, _ := strconv.Atoi(videoLengthStr)

	if err := validateContentSummary(obj, videoLength); err != nil {
		log.Printf("Content summary validation failed with the generated summary: %s", rawOutput)
		return "", err
	}

	obj.MediaUrl = fmt.Sprintf("https://storage.mtls.cloud.google.com/%s/%s", config.BasicRunConfig.InputBucket, config.BasicRunConfig.InputFile)
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal validated content summary: %w", err)
	}
	return string(objBytes), nil
}

func generatePrompt(config *common.GenaiStepConfig, videoLength string, contentType string) (string, error) {
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
	templateParams["VIDEO_LENGTH"] = videoLength
	templateParams["VIDEO_END_TIMESTAMP"] = convertSecondsToHHMMSS(videoLength)

	var buffer bytes.Buffer
	if err := config.GenaiRunConfig.TemplateService.GetTemplateBy(contentType).SummaryPrompt.Execute(&buffer, templateParams); err != nil {
		return "", err
	}

	return buffer.String(), nil

}

func convertIntSecondsToHHMMSS(intSeconds int) string {
	hours := intSeconds / 3600
	minutes := (intSeconds % 3600) / 60
	seconds := intSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func convertSecondsToHHMMSS(secondsStr string) string {
	s, err := strconv.Atoi(secondsStr)
	if err != nil {
		return ""
	}
	return convertIntSecondsToHHMMSS(s)
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

		//5. ensure the segment length is at minimum 5 second, shorter than this gemini won't give any valid response
		if end-start < 5 {
			return fmt.Errorf("segment %d: length of less than 5 seconds detected", i+1)
		}

		prevEnd = end
	}

	return nil
}

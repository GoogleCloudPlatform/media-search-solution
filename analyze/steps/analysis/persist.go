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
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
)

const (
	DefaultMovieTimeFormat = "15:04:05"
)

type InputObjects struct {
	ContentLength  int
	ContentSummary *model.MediaSummary
	Segments       []*model.Segment
}

func persist_analysis_result(genaiRunConfig *common.GenaiRunConfig) {
	stepConfig, err := common.NewGenaiStepConfig(common.PERSIST_STEP, genaiRunConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	stepConfig.StepLogic = persistResultLogicFunc(stepConfig)
	stepConfig.RunStep()
}

func persistResultLogicFunc(config *common.GenaiStepConfig) func() (string, error) {
	return func() (string, error) {
		inputObjects, err := getInputObjects(config.GenaiRunConfig)
		if err != nil {
			return "", err
		}

		for _, segment := range inputObjects.Segments {
			segment.Start = correctTimestamp(segment.Start, inputObjects.ContentLength)
			segment.End = correctTimestamp(segment.End, inputObjects.ContentLength)
		}

		sort.Slice(inputObjects.Segments, func(i, j int) bool {
			t, _ := time.Parse(DefaultMovieTimeFormat, inputObjects.Segments[i].Start)
			tt, _ := time.Parse(DefaultMovieTimeFormat, inputObjects.Segments[j].Start)
			return t.Before(tt)
		})

		return writeToBigQuery(config.GenaiRunConfig, createPersistObj(inputObjects))
	}
}

func writeToBigQuery(config *common.GenaiRunConfig, persistObj *model.Media) (string, error) {
	bqInserter := config.BigQueryClient.
		Dataset(config.CloudConfig.BigQueryDataSource.DatasetName).
		Table(config.CloudConfig.BigQueryDataSource.MediaTable).Inserter()
	if err := bqInserter.Put(config.Ctx, persistObj); err != nil {
		return "", err
	}
	return persistObj.Id, nil

}
func createPersistObj(inputObjects *InputObjects) *model.Media {
	media := model.NewMedia(inputObjects.ContentSummary.Title)
	media.Title = inputObjects.ContentSummary.Title
	media.Category = inputObjects.ContentSummary.Category
	media.Summary = inputObjects.ContentSummary.Summary
	media.MediaUrl = inputObjects.ContentSummary.MediaUrl
	media.LengthInSeconds = inputObjects.ContentLength
	media.Director = inputObjects.ContentSummary.Director
	media.ReleaseYear = inputObjects.ContentSummary.ReleaseYear
	media.Genre = inputObjects.ContentSummary.Genre
	media.Rating = inputObjects.ContentSummary.Rating
	media.Cast = append(media.Cast, inputObjects.ContentSummary.Cast...)
	media.Segments = append(media.Segments, inputObjects.Segments...)
	return media
}

func getInputObjects(config *common.GenaiRunConfig) (*InputObjects, error) {
	inputParameter := []string{
		common.CONTENT_LENGTH_STEP,
		common.CONTENT_SUMMARY_STEP,
	}
	inputValues := config.BasicRunConfig.GetStepsOutput(inputParameter)
	inpubObjects := &InputObjects{}
	inpubObjects.ContentLength, _ = strconv.Atoi(inputValues[common.CONTENT_LENGTH_STEP])
	contentSummary := &model.MediaSummary{}
	json.Unmarshal([]byte(inputValues[common.CONTENT_SUMMARY_STEP]), contentSummary)
	inpubObjects.ContentSummary = contentSummary

	segmentStepKeys := make([]string, len(contentSummary.SegmentTimeStamps))
	segments := make([]*model.Segment, 0)
	for i := range contentSummary.SegmentTimeStamps {
		segmentStepKeys[i] = getSegmentSummaryStepKey(i)
	}

	segmentStepValues := config.BasicRunConfig.GetStepsOutput(segmentStepKeys)

	for i, stepKey := range segmentStepKeys {
		segmentValue, ok := segmentStepValues[stepKey]
		if !ok {
			continue // Or handle missing segment summary
		}
		segment := &model.Segment{}
		json.Unmarshal([]byte(segmentValue), segment)
		if segment.SequenceNumber == 0 {
			log.Printf("Warning: Segment for step %s has sequence number 0. Overwriting with %d.", stepKey, i+1)
			segment.SequenceNumber = i + 1
		}
		if segment.Script == "" {
			log.Printf("Warning: Segment for step %s has empty script. Skipping.", stepKey)
			continue
		}
		segments = append(segments, segment)
	}
	inpubObjects.Segments = segments

	return inpubObjects, nil
}

func formatSeconds(totalSeconds int) string {
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// correctTimestamp attempts to fix malformed HH:MM:SS timestamps that are out of
// the video's duration range. It checks for a common LLM error where minutes
// are written as hours and seconds as minutes.
func correctTimestamp(timestampStr string, videoLength int) string {
	parts := strings.Split(timestampStr, ":")
	if len(parts) != 3 {
		return timestampStr
	}

	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])
	s, errS := strconv.Atoi(parts[2])

	if errH != nil || errM != nil || errS != nil {
		return timestampStr
	}

	originalSeconds := h*3600 + m*60 + s

	// If the timestamp is already valid, return it.
	if originalSeconds <= videoLength {
		return timestampStr
	}

	// The timestamp is out of bounds. Let's check for a common mix-up:
	// HH:MM:SS from the LLM should have been 00:HH:MM.
	correctedSeconds := h*60 + m
	if correctedSeconds <= videoLength {
		correctedTimestamp := fmt.Sprintf("00:%02d:%02d", h, m)
		return correctedTimestamp
	}

	// If correction is still out of bounds, clamp to video length as a last resort.
	clampedTimestamp := formatSeconds(videoLength)
	return clampedTimestamp
}

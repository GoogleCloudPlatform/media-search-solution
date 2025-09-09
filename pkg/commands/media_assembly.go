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
//         kingman (Charlie Wang)

package commands

import (
	"encoding/json"
	"fmt"

	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
)

const (
	DefaultMovieTimeFormat = "15:04:05"
)

type MediaAssembly struct {
	cor.BaseCommand
	summaryParam     string
	sceneParam       string
	mediaObjectParam string
	mediaLengthParam string
}

// NewMediaAssembly default constructor for MediaAssembly
func NewMediaAssembly(name string, summaryParam string, sceneParam string, mediaObjectParam string, mediaLengthParam string) *MediaAssembly {
	return &MediaAssembly{
		BaseCommand:      *cor.NewBaseCommand(name),
		summaryParam:     summaryParam,
		sceneParam:       sceneParam,
		mediaObjectParam: mediaObjectParam,
		mediaLengthParam: mediaLengthParam,
	}
}

// IsExecutable overrides the default to verify the summary param and scene param are in the context
func (m *MediaAssembly) IsExecutable(context cor.Context) bool {
	return context != nil &&
		context.Get(m.summaryParam) != nil &&
		context.Get(m.sceneParam) != nil
}

func (m *MediaAssembly) Execute(context cor.Context) {
	summary := context.Get(m.summaryParam).(*model.MediaSummary)
	jsonScenes := context.Get(m.sceneParam).([]string)
	mediaLengthInSeconds := context.Get(m.mediaLengthParam).(int)
	sceneValues := fmt.Sprintf("[ %s ]", strings.Join(jsonScenes, ","))

	scenes := make([]*model.Scene, 0)
	sceneErr := json.Unmarshal([]byte(sceneValues), &scenes)
	if sceneErr != nil {
		m.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(m.GetName(), sceneErr)
		return
	}

	if len(scenes) == 0 { // If no scenes were extracted, create a default scene with the summary.
		defaultScene := &model.Scene{
			SequenceNumber: 0,
			Start:          "00:00:00",
			End:            formatSeconds(mediaLengthInSeconds),
			Script:         summary.Summary,
		}
		scenes = append(scenes, defaultScene)
	}

	// Correct timestamps if they are out of bounds due to LLM mix-ups
	for _, scene := range scenes {
		scene.Start = correctTimestamp(scene.Start, mediaLengthInSeconds)
		scene.End = correctTimestamp(scene.End, mediaLengthInSeconds)
	}

	// Sort the scenes and sequence them
	sort.Slice(scenes, func(i, j int) bool {
		t, _ := time.Parse(DefaultMovieTimeFormat, scenes[i].Start)
		tt, _ := time.Parse(DefaultMovieTimeFormat, scenes[j].Start)
		return t.Before(tt)
	})
	for i, scene := range scenes {
		scene.SequenceNumber = i
	}

	// Call the constructor to ensure the UUID is generated
	// TODO - Base the
	media := model.NewMedia(summary.Title)
	media.Title = summary.Title
	media.Category = summary.Category
	media.Summary = summary.Summary
	media.MediaUrl = summary.MediaUrl
	media.LengthInSeconds = mediaLengthInSeconds
	media.Director = summary.Director
	media.ReleaseYear = summary.ReleaseYear
	media.Genre = summary.Genre
	media.Rating = summary.Rating
	media.Cast = append(media.Cast, summary.Cast...)
	media.Scenes = append(media.Scenes, scenes...)

	m.GetSuccessCounter().Add(context.GetContext(), 1)

	context.Add(m.mediaObjectParam, media)
	context.Add(cor.CtxOut, media)
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

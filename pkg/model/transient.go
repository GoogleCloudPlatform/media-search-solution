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

package model

// These objects are used in memory via workflows, but are not persisted to the dataset

// MediaFormatFilter is a simple video format object expressing the intended output
// and the destination width
type MediaFormatFilter struct {
	Format string
	Width  string
}

type TimeSpan struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type MediaSummary struct {
	Title             string        `json:"title"`
	Category          string        `json:"category"`
	Summary           string        `json:"summary"`
	LengthInSeconds   int           `json:"length_in_seconds"`
	MediaUrl          string        `json:"media_url,omitempty"`
	Director          string        `json:"director,omitempty"`
	ReleaseYear       int           `json:"release_year,omitempty"`
	Genre             string        `json:"genre,omitempty"`
	Rating            string        `json:"rating,omitempty"`
	Cast              []*CastMember `json:"cast,omitempty"`
	SegmentTimeStamps []*TimeSpan   `json:"segment_time_stamps,omitempty"`
}

type SegmentMatchResult struct {
	MediaId        string `json:"media_id" bigquery:"media_id"`
	SequenceNumber int    `json:"sequence_number" bigquery:"sequence_number"`
}

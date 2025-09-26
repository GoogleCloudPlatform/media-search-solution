// Copyright 2025 Google, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: rrmcguinness (Ryan McGuinness)
//         jaycherian (Jay Cherian)
//         kingman (Charlie Wang)

package model

import "google.golang.org/genai"

func NewMediaSummarySchema() *genai.Schema {
	// Define the schema for MediaSummary
	return &genai.Schema{
		Type: "object",
		Properties: map[string]*genai.Schema{
			"title":             {Type: "string"},
			"category":          {Type: "string"},
			"summary":           {Type: "string"},
			"length_in_seconds": {Type: "integer"},
			"media_url":         {Type: "string", Nullable: genai.Ptr(true)},
			"director":          {Type: "string", Nullable: genai.Ptr(true)},
			"release_year":      {Type: "integer", Nullable: genai.Ptr(true)},
			"genre":             {Type: "string", Nullable: genai.Ptr(true)},
			"rating":            {Type: "string", Nullable: genai.Ptr(true)},
			"cast": {
				Type:     "array",
				Nullable: genai.Ptr(true),
				Items: &genai.Schema{
					Type: "object",
					Properties: map[string]*genai.Schema{
						"character_name": {Type: "string"},
						"actor_name":     {Type: "string"},
					},
					Required: []string{"character_name", "actor_name"},
				},
			},
			"segment_time_stamps": {
				Type:     "array",
				Nullable: genai.Ptr(true),
				Items: &genai.Schema{
					Type: "object",
					Properties: map[string]*genai.Schema{
						"start": {Type: "string"},
						"end":   {Type: "string"},
					},
					Required: []string{"start", "end"},
				},
			},
		},
		Required: []string{"title", "category", "summary", "length_in_seconds"},
	}
}

func NewSegmentExtractorSchema() *genai.Schema {
	// Define the schema for SegmentExtractor
	return &genai.Schema{
		Type: "object",
		Properties: map[string]*genai.Schema{
			"sequence": {Type: "integer"},
			"start":    {Type: "string"},
			"end":      {Type: "string"},
			"script":   {Type: "string"},
		},
		Required: []string{"sequence", "start", "end", "script"},
	}
}

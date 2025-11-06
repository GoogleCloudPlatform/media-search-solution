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

import (
	"time"

	"github.com/google/uuid"
)

// Actor is used to represent the public details of an actor or actress.
type Actor struct {
	Id           string    `json:"id" bigquery:"id"`
	CreateDate   time.Time `json:"create_date" bigquery:"create_date"`
	Name         string    `json:"name" bigquery:"name"`
	DateOfBirth  time.Time `json:"dob" bigquery:"dob"` // Or time.Time if you need more precise handling
	DateOfDeath  time.Time `json:"dod" bigquery:"dod"`
	PlaceOfBirth string    `json:"pob" bigquery:"pob"`
	Biography    string    `json:"bio" bigquery:"bio"`
	Aliases      []string  `json:"aliases" bigquery:"aliases"`
	Awards       []string  `json:"awards" bigquery:"awards"`
	Nominations  []string  `json:"nominations" bigquery:"nominations"`
	ImageURL     string    `json:"ima_url" bigquery:"img_url"`
}

// Media capture the highest level of metadata about a media file.
type Media struct {
	Id              string        `json:"id" bigquery:"id"`
	CreateDate      time.Time     `json:"create_date" bigquery:"create_date"`
	Title           string        `json:"title" bigquery:"title"`
	Category        string        `json:"category" bigquery:"category"`
	Summary         string        `json:"summary" bigquery:"summary"`
	LengthInSeconds int           `json:"length_in_seconds" bigquery:"length_in_seconds"`
	MediaUrl        string        `json:"media_url" bigquery:"media_url"`
	Director        string        `json:"director,omitempty" bigquery:"director"`
	ReleaseYear     int           `json:"release_year,omitempty" bigquery:"release_year"`
	Genre           string        `json:"genre,omitempty" bigquery:"genre"`
	Rating          string        `json:"rating,omitempty" bigquery:"rating"`
	Cast            []*CastMember `json:"cast,omitempty" bigquery:"cast"`
	Segments        []*Segment    `json:"segments,omitempty" bigquery:"segments"`
}

func NewMedia(fileName string) *Media {
	// Use a UUID 5
	generatedID := uuid.NewSHA1(uuid.NameSpaceURL, ([]byte)(fileName))
	return &Media{
		Id:         generatedID.String(),
		CreateDate: time.Now(),
		Cast:       make([]*CastMember, 0),
		Segments:   make([]*Segment, 0),
	}
}

// Segment is a representation of a time span and it's sequence in a media object
// giving granular detail for the agent objects to interrogate
type Segment struct {
	SequenceNumber   int    `json:"sequence" bigquery:"sequence"`
	TokensToGenerate int    `json:"tokens_to_generate" bigquery:"tokens_to_generate"`
	TokensGenerated  int    `json:"tokens_generated" bigquery:"tokens_generated"`
	Start            string `json:"start" bigquery:"start"`
	End              string `json:"end" bigquery:"end"`
	Script           string `json:"script" bigquery:"script"`
}

// CastMember is a mapping object from a character to an actor
type CastMember struct {
	CharacterName string `json:"character_name" bigquery:"character_name"`
	ActorName     string `json:"actor_name" bigquery:"actor_name"`
}

// CastDialog is a mapping from a character to the spoken word in a segment
type CastDialog struct {
	CharacterName string `json:"character_name" bigquery:"character_name"`
	Dialog        string `json:"dialog" bigquery:"dialog"`
}

// SegmentEmbedding captures the summary embedding of a media file, good for general searches.
type SegmentEmbedding struct {
	Id             string    `json:"id" bigquery:"media_id"`
	SequenceNumber int       `json:"sequence_number" bigquery:"sequence_number"`
	ModelName      string    `json:"model_name" bigquery:"model_name"`
	Embeddings     []float64 `json:"embeddings" bigquery:"embeddings"`
}

func NewSegmentEmbedding(
	mediaId string,
	sequenceNumber int,
	modelName string) *SegmentEmbedding {

	return &SegmentEmbedding{
		Id:             mediaId,
		SequenceNumber: sequenceNumber,
		ModelName:      modelName,
		Embeddings:     make([]float64, 0),
	}
}

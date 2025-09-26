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

package model_test

import (
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewMedia(t *testing.T) {
	fileName := "test-file.mp4"
	media := model.NewMedia(fileName)

	// Use a UUID 5
	generatedID := uuid.NewSHA1(uuid.NameSpaceURL, ([]byte)(fileName))

	assert.Equal(t, generatedID.String(), media.Id)
	assert.WithinDuration(t, time.Now(), media.CreateDate, time.Second)
	assert.Equal(t, 0, len(media.Cast))
	assert.Equal(t, 0, len(media.Segments))
}

func TestNewSegmentEmbedding(t *testing.T) {
	mediaId := "test-media-id"
	sequenceNumber := 1
	modelName := "test-model"

	embedding := model.NewSegmentEmbedding(mediaId, sequenceNumber, modelName)

	assert.Equal(t, mediaId, embedding.Id)
	assert.Equal(t, sequenceNumber, embedding.SequenceNumber)
	assert.Equal(t, modelName, embedding.ModelName)
	assert.Equal(t, 0, len(embedding.Embeddings))
}

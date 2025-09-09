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

package workflow

import (
	"strings"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/commands"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
)

// DefaultFfmpegCommand The default command requires ffmpeg on the path of the running computer.
const DefaultFfmpegCommand = "ffmpeg"

// DefaultWidth The default width is the recommended size.
const DefaultWidth = "240"

type MediaResizeWorkflow struct {
	cor.BaseCommand
	ffmpegCommand    string
	videoFormat      *model.MediaFormatFilter
	storageClient    *storage.Client
	outputBucketName string
	chain            cor.Chain
	config           *cloud.Config
}

func (m *MediaResizeWorkflow) Execute(context cor.Context) {
	m.chain.Execute(context)
}

func (m *MediaResizeWorkflow) initializeChain() {
	out := cor.NewBaseChain(m.GetName())

	// Convert the Message to an Object
	out.AddCommand(commands.NewMediaTriggerToGCSObject("gcs-topic-listener"))

	// Run FFMpeg
	out.AddCommand(commands.NewFFMpegCommand("video-resize", m.ffmpegCommand, m.videoFormat.Width, m.config))

	m.chain = out
}

func NewMediaResizeWorkflow(
	config *cloud.Config,
	serviceClients *cloud.ServiceClients,
	ffmpegCommand string,
	videoFormat *model.MediaFormatFilter) *MediaResizeWorkflow {

	// Ensure the FFMPegCommand is set, otherwise use the default
	if len(strings.Trim(ffmpegCommand, " ")) == 0 {
		ffmpegCommand = DefaultFfmpegCommand
	}

	// Set the default width
	if videoFormat == nil {
		videoFormat = &model.MediaFormatFilter{Width: DefaultWidth, Format: "mp4"}
	}

	out := &MediaResizeWorkflow{
		BaseCommand:      *cor.NewBaseCommand("media-resize-workflow"),
		ffmpegCommand:    ffmpegCommand,
		videoFormat:      videoFormat,
		storageClient:    serviceClients.StorageClient,
		config:           config,
		outputBucketName: config.Storage.LowResOutputBucket}
	out.initializeChain()
	return out
}

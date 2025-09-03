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

package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/GoogleCloudPlatform/solutions/media/pkg/cloud"
	"github.com/GoogleCloudPlatform/solutions/media/pkg/cor"
)

type AudioExtractorCommand struct {
	cor.BaseCommand
	commandPath string
	config      *cloud.Config
}

const (
	DefaultAudioExtractorCmdArgs = "-i %s -q:a 0 -map a -ac 1 %s"
)

func NewAudioExtractorCommand(name string, commandPath string, config *cloud.Config) *AudioExtractorCommand {
	out := AudioExtractorCommand{
		BaseCommand: *cor.NewBaseCommand(name),
		commandPath: commandPath,
		config:      config,
	}
	return &out
}

func (c *AudioExtractorCommand) Execute(context cor.Context) {
	gcsFile := context.Get(cloud.GetGCSObjectName()).(*cloud.GCSObject)
	inputFileName := fmt.Sprintf("%s/%s/%s", c.config.Storage.GCSFuseMountPoint, gcsFile.Bucket, gcsFile.Name)

	outputFileName := strings.TrimSuffix(gcsFile.Name, ".mp4") + ".wav"

	outputFileFullPath := fmt.Sprintf("%s/%s/%s", c.config.Storage.GCSFuseMountPoint, c.config.Storage.AudioBucket, outputFileName)

	args := fmt.Sprintf(DefaultAudioExtractorCmdArgs, inputFileName, outputFileFullPath)

	cmd := exec.Command(c.commandPath, strings.Split(args, CommandSeparator)...)
	cmd.Stderr = os.Stderr
	_, err := cmd.Output()

	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("error extracting audio with ffmpeg: %w", err))
		return
	}
	c.GetSuccessCounter().Add(context.GetContext(), 1)
	context.Add(c.GetOutputParam(), fmt.Sprintf("gs://%s/%s", c.config.Storage.AudioBucket, outputFileName))
}

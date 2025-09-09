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
// Author: kingman (Charlie Wang)

package commands

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
)

const (
	DefaultVideoDurationCmdArgs = "-v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 %s"
	// FileCheckRetries is the number of times to check for a file's existence.
	FileCheckRetries = 5
	// FileCheckDelay is the time to wait between file existence checks.
	FileCheckDelay = 10 * time.Second
)

type MediaLengthCommand struct {
	cor.BaseCommand
	commandPath string
	config      *cloud.Config
}

func NewMediaLengthCommand(name string, commandPath string, outputParamName string, config *cloud.Config) *MediaLengthCommand {
	out := MediaLengthCommand{
		BaseCommand: *cor.NewBaseCommand(name),
		commandPath: commandPath,
		config:      config,
	}
	out.OutputParamName = outputParamName
	return &out
}

func (c *MediaLengthCommand) Execute(context cor.Context) {
	gcsFile := context.Get(cloud.GetGCSObjectName()).(*cloud.GCSObject)
	inputFileName := fmt.Sprintf("%s/%s/%s", c.config.Storage.GCSFuseMountPoint, gcsFile.Bucket, gcsFile.Name)
	log.Printf("Received message for media file: %s/%s", gcsFile.Bucket, gcsFile.Name)

	var err error
	for i := range FileCheckRetries {
		if _, err = os.Stat(inputFileName); err == nil {
			break
		}
		log.Printf("waiting for file to appear: %s, attempt %d/%d", inputFileName, i+1, FileCheckRetries)
		time.Sleep(FileCheckDelay)
	}

	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("file: %s not found after several retries. Error: %w", inputFileName, err))
		return
	}

	args := fmt.Sprintf(DefaultVideoDurationCmdArgs, inputFileName)
	cmd := exec.Command(c.commandPath, strings.Split(args, CommandSeparator)...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("error running ffprobe: %w", err))
		return
	}

	length, err := extractVideoLengthToFullSeconds(output)
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), err)
		return
	}
	c.GetSuccessCounter().Add(context.GetContext(), 1)

	context.Add(c.GetOutputParam(), length)
	context.Add(cor.CtxOut, length)
}

func extractVideoLengthToFullSeconds(output []byte) (int, error) {
	s := strings.TrimSpace(string(output))

	duration, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return int(duration) + 1, nil
	}
	return 0, fmt.Errorf("got invalid video duration: %s", s)
}

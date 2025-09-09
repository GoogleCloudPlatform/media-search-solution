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
	"fmt"
	"io"
	"log"
	"path/filepath"
	"time"

	"os"
	"os/exec"
	"strings"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
)

const (
	DefaultFfmpegArgs = "-analyzeduration 0 -probesize 5000000 -y -hide_banner -i %s -filter:v scale=w=%s:h=trunc(ow/a/2)*2 -f mp4 %s"
	TempFilePrefix    = "ffmpeg-output-"
	CommandSeparator  = " "
)

// FFMpegCommand is a simple command used for
// downloading a media file embedded in the message, resizing it
// and uploading the resized version to the destination bucket.
// The scale uses a dynamic scale to keep the aspect ratio of the original.
type FFMpegCommand struct {
	cor.BaseCommand
	commandPath string
	targetWidth string
	config      *cloud.Config
}

func NewFFMpegCommand(name string, commandPath string, targetWidth string, config *cloud.Config) *FFMpegCommand {
	return &FFMpegCommand{
		BaseCommand: *cor.NewBaseCommand(name),
		commandPath: commandPath,
		targetWidth: targetWidth,
		config:      config}
}

// Execute executes the business logic of the command
func (c *FFMpegCommand) Execute(context cor.Context) {
	msg := context.Get(c.GetInputParam()).(*cloud.GCSObject)
	inputFileName := fmt.Sprintf("%s/%s/%s", c.config.Storage.GCSFuseMountPoint, msg.Bucket, msg.Name)
	log.Printf("Received message for media file: %s/%s", msg.Bucket, msg.Name)

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

	file, err := os.Open(inputFileName)
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), err)
		return
	}
	tempFile, _ := os.CreateTemp("", TempFilePrefix)

	args := fmt.Sprintf(DefaultFfmpegArgs, file.Name(), c.targetWidth, tempFile.Name())
	cmd := exec.Command(c.commandPath, strings.Split(args, CommandSeparator)...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("error running ffmpeg: %w", err))
		return
	}

	outputName := msg.Name
	if ext := filepath.Ext(outputName); ext != ".mp4" {
		outputName = strings.TrimSuffix(outputName, ext) + ".mp4"
	}

	outputFile := fmt.Sprintf("%s/%s/%s", c.config.Storage.GCSFuseMountPoint, c.config.Storage.LowResOutputBucket, outputName)

	MoveFile(tempFile.Name(), outputFile)
	c.GetSuccessCounter().Add(context.GetContext(), 1)
	context.AddTempFile(outputFile)
	context.Add(cor.CtxOut, outputFile)
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("could not open source file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("could not open dest file: %v", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("could not copy to dest from source: %v", err)
	}

	inputFile.Close()

	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("could not remove source file: %v", err)
	}
	return nil
}

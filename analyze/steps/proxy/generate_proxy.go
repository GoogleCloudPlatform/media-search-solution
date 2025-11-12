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
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	common "github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
)

const (
	DefaultArgsStringTemplate = "-analyzeduration 0 -probesize 5000000 -y -hide_banner -i %s -filter:v scale=w=%s:h=trunc(ow/a/2)*2 -f mp4 %s"
	TempFilePrefix            = "ffmpeg-output-"
)

type ProxyCommandConfig struct {
	common.CommandStepConfig
	OutputFolder string
	TargetWidth  string
	OutputFormat string
}

func NewProxyCommandConfig(basicRunConfig *common.BasicRunConfig, stepKey, commandPath, argsStringTemplate, outputFolder, targetWidth, outputFormat string) *ProxyCommandConfig {
	commandStepConfig := common.NewCommandStepConfig(basicRunConfig, stepKey, commandPath, argsStringTemplate, nil)
	config := &ProxyCommandConfig{
		CommandStepConfig: *commandStepConfig,
		OutputFolder:      outputFolder,
		TargetWidth:       targetWidth,
		OutputFormat:      outputFormat,
	}
	commandStepConfig.CommandLogic = config.proxyStepLogic
	return config
}

func (config *ProxyCommandConfig) proxyStepLogic(inputFileFullPath string) (string, error) {
	file, err := os.Open(inputFileFullPath)
	if err != nil {
		log.Fatalf("error opening input file %s: %v", inputFileFullPath, err)
	}
	tempFile, _ := os.CreateTemp("", TempFilePrefix)

	args := fmt.Sprintf(config.ArgsStringTemplate, file.Name(), config.TargetWidth, tempFile.Name())
	cmd := exec.Command(config.CommandPath, strings.Split(args, common.CommandSeparator)...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("error running ffmpeg command: %v", err)
	}

	outputName := config.BasicRunConfig.InputFile
	if ext := filepath.Ext(config.BasicRunConfig.InputFile); ext != config.OutputFormat {
		outputName = strings.TrimSuffix(outputName, ext) + config.OutputFormat
	}

	outputFile := fmt.Sprintf("%s/%s/%s", config.BasicRunConfig.MountPoint, config.OutputFolder, outputName)

	err = MoveFile(tempFile.Name(), outputFile)
	if err != nil {
		log.Fatalf("error moving file: %v", err)
	}

	return fmt.Sprintf("%s/%s", config.OutputFolder, outputName), nil
}

func main() {
	commandPath := common.Getenv("COMMAND_PATH", "bin/ffmpeg")
	outputFormat := common.Getenv("OUTPUT_FORMAT", ".mp4")
	targetWidth := common.Getenv("OUTPUT_WIDTH", "240")
	outputFolder := os.Getenv("OUTPUT_FOLDER")
	if len(outputFolder) == 0 {
		log.Fatal("OUTPUT_FOLDER not specified")
	}
	basicRunConfig, err := common.NewBasicRunConfig()
	if err != nil {
		log.Fatal(err)
	}
	config := NewProxyCommandConfig(basicRunConfig, common.GENERATE_PROXY_STEP, commandPath, DefaultArgsStringTemplate, outputFolder, targetWidth, outputFormat)

	config.RunStep()
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

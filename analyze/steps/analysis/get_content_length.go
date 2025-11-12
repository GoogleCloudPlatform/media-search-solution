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
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	common "github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
)

const (
	ContentLengthCmdArgs = "-v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 %s"
)

type ContentLengthCommandConfig struct {
	common.CommandStepConfig
	CommandPath string
}

func NewContentLengthCommandConfig(basicRunConfig *common.BasicRunConfig, stepKey, commandPath, argsStringTemplate string) *ContentLengthCommandConfig {
	commandStepConfig := common.NewCommandStepConfig(basicRunConfig, stepKey, commandPath, argsStringTemplate, nil)
	config := &ContentLengthCommandConfig{
		CommandStepConfig: *commandStepConfig,
		CommandPath:       commandPath,
	}
	commandStepConfig.CommandLogic = config.contentLengthStepLogic
	return config
}

func (config *ContentLengthCommandConfig) contentLengthStepLogic(inputFileFullPath string) (string, error) {
	args := fmt.Sprintf(config.ArgsStringTemplate, inputFileFullPath)
	cmd := exec.Command(config.CommandPath, strings.Split(args, common.CommandSeparator)...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("error running ffprobe: %v", err)
	}
	content_length, err := extractVideoLengthToFullSeconds(output)
	if err != nil {
		log.Fatalf("error extracting video length: %v", err)
	}
	return strconv.Itoa(content_length), nil
}

func get_content_length(basicRunConfig *common.BasicRunConfig) {
	commandPath := common.Getenv("COMMAND_PATH", "bin/ffprobe")

	config := NewContentLengthCommandConfig(basicRunConfig, common.CONTENT_LENGTH_STEP, commandPath, ContentLengthCmdArgs)

	config.RunStep()
}

func extractVideoLengthToFullSeconds(output []byte) (int, error) {
	s := strings.TrimSpace(string(output))

	duration, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return int(duration) + 1, nil
	}
	return 0, fmt.Errorf("got invalid video duration: %s", s)
}

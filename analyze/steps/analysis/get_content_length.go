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

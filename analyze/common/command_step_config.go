package common

import (
	"fmt"
	"os"
	"time"
)

const (
	FileCheckRetries = 2
	FileCheckDelay   = 2 * time.Second
	CommandSeparator = " "
)

type CommandStepConfig struct {
	BasicStepConfig
	CommandPath        string
	ArgsStringTemplate string
	CommandLogic       func(string) (string, error)
}

func NewCommandStepConfig(basicRunConfig *BasicRunConfig, stepKey, commandPath, argsStringTemplate string, commandLogic func(string) (string, error)) *CommandStepConfig {
	config := &CommandStepConfig{
		BasicStepConfig: BasicStepConfig{
			BasicRunConfig: basicRunConfig,
			StepKey:        stepKey,
		},
		CommandPath:        commandPath,
		ArgsStringTemplate: argsStringTemplate,
		CommandLogic:       commandLogic,
	}
	config.StepLogic = config.commandStepLogic
	return config
}

func (config *CommandStepConfig) commandStepLogic() (string, error) {
	inputFileFullPath, err := config.waitForInputFile()
	if err != nil {
		return "", err
	}
	return config.CommandLogic(inputFileFullPath)
}

func (config *CommandStepConfig) waitForInputFile() (string, error) {
	inputFileFullPath := config.BasicRunConfig.MountPoint + "/" + config.BasicRunConfig.InputBucket + "/" + config.BasicRunConfig.InputFile
	for range FileCheckRetries {
		if _, err := os.Stat(inputFileFullPath); err == nil {
			return inputFileFullPath, nil
		}
		time.Sleep(FileCheckDelay)
	}
	return "", fmt.Errorf("input file %s not found after %d retries", inputFileFullPath, FileCheckRetries)

}

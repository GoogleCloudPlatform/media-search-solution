package common

import (
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
)

type BasicStepConfig struct {
	BasicRunConfig *BasicRunConfig
	StepKey        string
	StepLogic      func() (string, error)
}

func NewBasicStepConfig(basicRunConfig *BasicRunConfig, stepKey string, stepLogic func() (string, error)) *BasicStepConfig {
	return &BasicStepConfig{
		BasicRunConfig: basicRunConfig,
		StepKey:        stepKey,
		StepLogic:      stepLogic,
	}
}

func (config *BasicStepConfig) StepCompleted() bool {
	status := config.BasicRunConfig.GetStepStatusByKey(config.StepKey)
	if status != nil && status.Status == StepCompleted {
		return true
	}
	return false
}

func (config *BasicStepConfig) setStepStatusToCompleted(output string) (string, error) {
	status := StepStatus{
		Output: output,
		Status: StepCompleted,
	}
	statusBytes, err := json.Marshal(status)
	if err != nil {
		return "", err
	}
	if stepKey, err := config.UpdateGCSObjectMetadata(
		map[string]string{
			config.StepKey: string(statusBytes),
		}); err != nil {
		return "", err
	} else {
		return stepKey, err
	}
}

func (config *BasicStepConfig) UpdateGCSObjectMetadata(metadata map[string]string) (string, error) {
	storageClient, err := storage.NewClient(config.BasicRunConfig.Ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %v", err)
	}
	defer storageClient.Close()
	objectUpdate := storage.ObjectAttrsToUpdate{
		Metadata: metadata,
	}
	obj := storageClient.Bucket(config.BasicRunConfig.InputBucket).Object(config.BasicRunConfig.InputFile)
	if _, err := obj.Update(config.BasicRunConfig.Ctx, objectUpdate); err != nil {
		return "", fmt.Errorf("failed to update object metadata: %v", err)
	}
	return config.StepKey, nil
}

func (config *BasicStepConfig) RunStep() {
	if config.StepCompleted() {
		log.Printf("Step %s already completed for %s/%s, skipping step", config.StepKey, config.BasicRunConfig.InputBucket, config.BasicRunConfig.InputFile)
		return
	}

	output, err := config.StepLogic()

	if err != nil {
		log.Fatalf("error executing step %s: %v", config.StepKey, err)
	}

	if _, err := config.setStepStatusToCompleted(output); err != nil {
		log.Fatalf("error setting step %s status to completed: %v", config.StepKey, err)
	}
}

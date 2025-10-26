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

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

type BasicRunConfig struct {
	InputFile     string
	InputBucket   string
	MountPoint    string
	Ctx           context.Context
	storageClient *storage.Client
}

func NewBasicRunConfig() (*BasicRunConfig, error) {
	inputBucket, inputFile, err := parseInputFileEnv()
	if err != nil {
		return nil, err
	}

	mountPoint := Getenv("MOUNT_POINT", "/mnt")
	return &BasicRunConfig{
		InputBucket: inputBucket,
		InputFile:   inputFile,
		MountPoint:  mountPoint,
		Ctx:         context.Background(),
	}, nil
}

func parseInputFileEnv() (string, string, error) {
	inputFile := os.Getenv("INPUT_FILE")
	if len(inputFile) == 0 {
		return "", "", fmt.Errorf("no input file specified")
	}

	parts := strings.SplitN(inputFile, "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid INPUT_FILE format: expected 'bucket/object', got '%s'", inputFile)
	}

	return parts[0], parts[1], nil
}

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func (config *BasicRunConfig) getRunSourceObjectMetadata() (map[string]string, error) {
	storageClient := config.GetStorageClient()

	bucket := storageClient.Bucket(config.InputBucket)
	obj := bucket.Object(config.InputFile)

	attrs, err := obj.Attrs(config.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %v", err)
	}

	return attrs.Metadata, nil
}

func (config *BasicRunConfig) GetStorageClient() *storage.Client {
	if config.storageClient == nil {
		config.storageClient, _ = storage.NewClient(config.Ctx)
	}
	return config.storageClient
}

func (config *BasicRunConfig) SetStorageClient(storageClient *storage.Client) {
	if config.storageClient != nil {
		config.storageClient.Close()
	}
	config.storageClient = storageClient
}

func (config *BasicRunConfig) GetStepStatusByKey(stepKey string) *StepStatus {
	metadata, err := config.getRunSourceObjectMetadata()
	if err != nil {
		return nil
	}
	val, ok := metadata[stepKey]
	if !ok {
		return nil
	}

	var status StepStatus
	if err := json.Unmarshal([]byte(val), &status); err != nil {
		return nil
	}
	return &status
}

func (config *BasicRunConfig) getStepsField(stepKeys []string, extractor func(status *StepStatus) string) map[string]string {
	outputs := make(map[string]string)
	metadata, err := config.getRunSourceObjectMetadata()
	if err != nil {
		return outputs
	}
	for _, stepKey := range stepKeys {
		val, ok := metadata[stepKey]
		if !ok {
			continue
		}

		var status StepStatus
		if err := json.Unmarshal([]byte(val), &status); err != nil {
			continue
		}
		outputs[stepKey] = extractor(&status)
	}
	return outputs
}

func (config *BasicRunConfig) GetStepsStatus(stepKeys []string) map[string]string {
	return config.getStepsField(stepKeys, func(status *StepStatus) string { return status.Status })
}

func (config *BasicRunConfig) GetStepsOutput(stepKeys []string) map[string]string {
	return config.getStepsField(stepKeys, func(status *StepStatus) string { return status.Output })
}

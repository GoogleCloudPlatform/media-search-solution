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
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/metric"
	"google.golang.org/genai"
)

type GenAICounter struct {
	InputCounter  metric.Int64Counter
	OutputCounter metric.Int64Counter
	RetryCounter  metric.Int64Counter
}

type GenaiStepConfig struct {
	BasicStepConfig
	GenaiRunConfig *GenaiRunConfig
	Counters       *GenAICounter
}

func NewGenaiStepConfig(stepKey string, genaiRunConfig *GenaiRunConfig, stepLogic func() (string, error)) (*GenaiStepConfig, error) {
	basicStepConfig := NewBasicStepConfig(&genaiRunConfig.BasicRunConfig, stepKey, stepLogic)
	inputCounter, _ := genaiRunConfig.Meter.Int64Counter(stepKey + ".gemini.token.input")
	outputCounter, _ := genaiRunConfig.Meter.Int64Counter(stepKey + ".gemini.token.output")
	retryCounter, _ := genaiRunConfig.Meter.Int64Counter(stepKey + ".gemini.token.retry")
	counters := &GenAICounter{
		InputCounter:  inputCounter,
		OutputCounter: outputCounter,
		RetryCounter:  retryCounter,
	}
	return &GenaiStepConfig{
		BasicStepConfig: *basicStepConfig,
		GenaiRunConfig:  genaiRunConfig,
		Counters:        counters,
	}, nil
}

func getContentCacheMetaDataKey(modelName string, systemInstructionCacheId string) string {
	return fmt.Sprintf("%s_%s_%s", "ims_genai_cache", modelName, systemInstructionCacheId)
}

func getContentCacheMetaDataKeyWithChunk(modelName string, systemInstructionCacheId string, startOffsetSec int, endOffsetSec int) string {
	return fmt.Sprintf("%s_%s_%s_%d_%d", "ims_genai_cache", modelName, systemInstructionCacheId, startOffsetSec, endOffsetSec)
}

func getSystemInstructionCacheId(systemInstruction *genai.Content) string {
	if systemInstruction == nil || len(systemInstruction.Parts) == 0 {
		return "no_system_instruction"
	}
	var instructionText string
	for _, part := range systemInstruction.Parts {
		instructionText += part.Text
	}
	checksum := crc32.ChecksumIEEE([]byte(instructionText))
	return strconv.FormatUint(uint64(checksum), 16)
}

func (config *GenaiStepConfig) loadGenaiContentCacheFromMetadata(cacheMetaDataKey string) *genai.CachedContent {
	cacheStatus := config.BasicRunConfig.GetStepStatusByKey(cacheMetaDataKey)
	if cacheStatus != nil && cacheStatus.Status == StepCompleted {
		var cachedContent genai.CachedContent
		if err := json.Unmarshal([]byte(cacheStatus.Output), &cachedContent); err == nil {
			if cachedContent.ExpireTime.After(time.Now()) {
				log.Printf("Reuse cache: %s", cachedContent.Name)
				return &cachedContent
			}
		}
	}
	return nil
}

func (config *GenaiStepConfig) createGenaiContentCache(modelName string, contents []*genai.Content, systemInstruction *genai.Content) (*genai.CachedContent, error) {
	model := config.GenaiRunConfig.AgentModels[modelName]
	return config.GenaiRunConfig.GenAIClient.Caches.Create(config.BasicRunConfig.Ctx, model.ModelName, &genai.CreateCachedContentConfig{
		Contents:          contents,
		SystemInstruction: systemInstruction,
	})
}

func (config *GenaiStepConfig) persistGenaiContentCache(cachedContent *genai.CachedContent, cacheMetaDataKey string) {
	contentCacheStr, err := json.Marshal(cachedContent)
	if err != nil {
		return
	}
	status := StepStatus{
		Output: string(contentCacheStr),
		Status: StepCompleted,
	}
	statusBytes, err := json.Marshal(status)
	if err != nil {
		return
	}
	if _, err := config.UpdateGCSObjectMetadata(
		map[string]string{
			cacheMetaDataKey: string(statusBytes),
		}); err != nil {
		return
	}
}

func (config *GenaiStepConfig) GetGenaiContentCacheWithChunk(modelName string, systemInstruction *genai.Content, startOffsetSec int, endOffsetSec int) (*genai.CachedContent, error) {
	systemInstructionCacheId := getSystemInstructionCacheId(systemInstruction)
	cacheMetaDataKey := getContentCacheMetaDataKeyWithChunk(modelName, systemInstructionCacheId, startOffsetSec, endOffsetSec)
	cachedContent := config.loadGenaiContentCacheFromMetadata(cacheMetaDataKey)
	if cachedContent != nil {
		return cachedContent, nil
	}

	startOffset := time.Duration(startOffsetSec) * time.Second
	endOffset := time.Duration(endOffsetSec) * time.Second

	contents := []*genai.Content{
		{Parts: []*genai.Part{
			{
				FileData: &genai.FileData{
					FileURI:  config.GenaiRunConfig.GetInputFileGCSURI(),
					MIMEType: GENAI_INPUT_FILE_TYPE,
				},
				VideoMetadata: &genai.VideoMetadata{
					StartOffset: startOffset,
					EndOffset:   endOffset,
				},
			},
		},
			Role: genai.RoleUser},
	}

	genaiContentCache, err := config.createGenaiContentCache(modelName, contents, systemInstruction)
	if err != nil {
		return nil, err
	}

	config.persistGenaiContentCache(genaiContentCache, cacheMetaDataKey)

	return genaiContentCache, nil
}

func (config *GenaiStepConfig) GetGenaiContentCache(modelName string, systemInstruction *genai.Content) (*genai.CachedContent, error) {
	systemInstructionCacheId := getSystemInstructionCacheId(systemInstruction)
	cacheMetaDataKey := getContentCacheMetaDataKey(modelName, systemInstructionCacheId)
	cachedContent := config.loadGenaiContentCacheFromMetadata(cacheMetaDataKey)
	if cachedContent != nil {
		return cachedContent, nil
	}

	gcsFileLink := config.GenaiRunConfig.GetInputFileGCSURI()
	contents := []*genai.Content{
		{Parts: []*genai.Part{
			genai.NewPartFromURI(gcsFileLink, GENAI_INPUT_FILE_TYPE),
		},
			Role: "user"},
	}

	genaiContentCache, err := config.createGenaiContentCache(modelName, contents, systemInstruction)
	if err != nil {
		return nil, err
	}

	config.persistGenaiContentCache(genaiContentCache, cacheMetaDataKey)

	return genaiContentCache, nil
}

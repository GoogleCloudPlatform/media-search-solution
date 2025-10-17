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
	"log"
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

func getContentCacheMetaDataKey(modelName string, stepCacheId string) string {
	return fmt.Sprintf("%s_%s_%s", "ims_genai_cache", modelName, stepCacheId)
}

func (config *GenaiStepConfig) GetGenaiContentCache(modelName string, stepCacheId string, systemInstruction *genai.Content) (*genai.CachedContent, error) {
	cacheMetaDataKey := getContentCacheMetaDataKey(modelName, stepCacheId)

	cacheStatus := config.BasicRunConfig.GetStepStatusByKey(cacheMetaDataKey)
	if cacheStatus != nil && cacheStatus.Status == StepCompleted {

		var cachedContent genai.CachedContent
		if err := json.Unmarshal([]byte(cacheStatus.Output), &cachedContent); err == nil {
			if cachedContent.ExpireTime.After(time.Now()) {
				log.Printf("Reuse cache: %s", cachedContent.Name)
				return &cachedContent, nil
			}
		}
	}

	gcsFileLink := config.GenaiRunConfig.GetInputFileGCSURI()
	contents := []*genai.Content{
		{Parts: []*genai.Part{
			genai.NewPartFromURI(gcsFileLink, GENAI_INPUT_FILE_TYPE),
		},
			Role: "user"},
	}
	model := config.GenaiRunConfig.AgentModels[modelName]
	genaiContentCache, err := config.GenaiRunConfig.GenAIClient.Caches.Create(config.BasicRunConfig.Ctx, model.ModelName, &genai.CreateCachedContentConfig{
		Contents:          contents,
		SystemInstruction: systemInstruction,
	})
	if err != nil {
		return nil, err
	}

	contentCacheStr, err := json.Marshal(genaiContentCache)
	if err != nil {
		return nil, err
	}
	status := StepStatus{
		Output: string(contentCacheStr),
		Status: StepCompleted,
	}
	statusBytes, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}
	if _, err := config.UpdateGCSObjectMetadata(
		map[string]string{
			cacheMetaDataKey: string(statusBytes),
		}); err != nil {
		return nil, err
	}

	return genaiContentCache, nil

}

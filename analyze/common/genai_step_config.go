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
	"go.opentelemetry.io/otel/metric"
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

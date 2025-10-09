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

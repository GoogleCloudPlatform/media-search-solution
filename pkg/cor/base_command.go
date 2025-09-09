// Copyright 2024 Google, LLC
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

package cor

import (
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// BaseCommand is the default implementation of Command
type BaseCommand struct {
	Name            string
	InputParamName  string
	OutputParamName string
	Tracer          trace.Tracer
	Meter           metric.Meter
	SuccessCounter  metric.Int64Counter
	ErrorCounter    metric.Int64Counter
}

func NewBaseCommand(name string) *BaseCommand {
	meter := otel.Meter("github.com/GoogleCloudPlatform/media-search-solution")
	successCounter, err := meter.Int64Counter(fmt.Sprintf("%s.counter.success", name))
	if err != nil {
		log.Printf("error creating success counter: %s\n", name)
	}
	errorCounter, err := meter.Int64Counter(fmt.Sprintf("%s.counter.error", name))
	if err != nil {
		log.Printf("error creating error counter: %s\n", name)
	}
	return &BaseCommand{
		Name:           name,
		Tracer:         otel.Tracer(name),
		Meter:          meter,
		SuccessCounter: successCounter,
		ErrorCounter:   errorCounter,
	}
}

func (c *BaseCommand) GetName() string {
	return c.Name
}

// IsExecutable a default implementation of IsExecutable.
func (c *BaseCommand) IsExecutable(context Context) bool {
	return context != nil && context.Get(c.GetInputParam()) != nil && context.GetContext() != nil
}

// GetInputParam the name of the parameter expected as the primary input,
// if empty it will default to CtxIn, during a chain execution event CtxIn
// will be mapped to the previous executions CtxOut to ensure PIPE / chain behaviors.
func (c *BaseCommand) GetInputParam() string {
	if len(c.InputParamName) == 0 {
		return CtxIn
	}
	return c.InputParamName
}

// GetOutputParam the name of the output parameter, the default is CtxOut
// See the chain execute method for more detail.
func (c *BaseCommand) GetOutputParam() string {
	if len(c.OutputParamName) == 0 {
		return CtxOut
	}
	return c.OutputParamName
}

func (c *BaseCommand) GetTracer() trace.Tracer {
	return c.Tracer
}

func (c *BaseCommand) GetMeter() metric.Meter {
	return c.Meter
}

func (c *BaseCommand) GetSuccessCounter() metric.Int64Counter {
	return c.SuccessCounter
}

func (c *BaseCommand) GetErrorCounter() metric.Int64Counter {
	return c.ErrorCounter
}

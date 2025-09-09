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
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	CtxIn  = "__IN__"
	CtxOut = "__OUT__"
)

// Context is an opinionated runtime context for Go Lang.
// It's a bit more complex than other language versions due to the nature
// of Filesystem behaviors.
type Context interface {
	SetContext(context context.Context)
	GetContext() context.Context
	Add(key string, value interface{}) Context
	AddError(key string, err error)
	GetErrors() map[string]error
	Get(key string) interface{}
	Remove(key string)
	HasErrors() bool
	AddTempFile(file string)
	GetTempFiles() []string
	Close()
}

type Executable interface {
	Execute(context Context)
}

// Command is a simple interface that ensures an atomic unit of work.
// The principals of a Command are: 1) Atomic, 2) Testable, and 3) Thread Safe
type Command interface {
	Executable
	GetName() string
	GetInputParam() string
	GetOutputParam() string
	IsExecutable(context Context) bool
	GetTracer() trace.Tracer
	GetMeter() metric.Meter
	GetSuccessCounter() metric.Int64Counter
	GetErrorCounter() metric.Int64Counter
}

// Chain is a collection of commands that ensure the serial or parallel execution
// of the commands. The Chain is a command and therefore inherits the principals of the command
// and in addition each Chain implements it's own execution strategy.
type Chain interface {
	Command
	ContinueOnFailure(bool) Chain
	AddCommand(command Command) Chain
}

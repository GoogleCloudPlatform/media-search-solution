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

	"go.opentelemetry.io/otel/codes"
)

type BaseChain struct {
	BaseCommand
	continueOnFailure bool
	commands          []Command
}

func NewBaseChain(name string) *BaseChain {
	return &BaseChain{BaseCommand: *NewBaseCommand(name)}
}

func (c *BaseChain) ContinueOnFailure(continueOnFailure bool) Chain {
	c.continueOnFailure = continueOnFailure
	return c
}

func (c *BaseChain) AddCommand(command Command) Chain {
	c.commands = append(c.commands, command)
	return c
}

func (c *BaseChain) GetCommands() []Command {
	return c.commands
}

func (c *BaseChain) IsExecutable(context Context) bool {
	return context.GetContext() != nil
}

func (c *BaseChain) Execute(chCtx Context) {
	var ctx = chCtx.GetContext()
	var parentCtx = chCtx.GetContext()

	outerCtx, chainSpan := c.Tracer.Start(ctx, fmt.Sprintf("%s_execute", c.GetName()))
	for _, command := range c.commands {
		// Ensure that the next parameter is callable in a pipe stack
		commandContext, commandSpan := c.Tracer.Start(outerCtx, command.GetName())
		commandSpan.SetName(command.GetName())
		if chCtx.HasErrors() && !c.continueOnFailure {
			commandSpan.SetStatus(codes.Error, "previous error on chain")
			break
		} else if command.IsExecutable(chCtx) {
			// Since the next command may be a chain, we must set the parent context
			chCtx.SetContext(commandContext)

			// Start a span for each command to measure command performance
			command.Execute(chCtx)

			// Reset the context to the original state
			if parentCtx != nil {
				chCtx.SetContext(parentCtx)
			} else {
				chCtx.SetContext(nil)
			}
		} else {
			commandSpan.SetStatus(codes.Error, fmt.Sprintf("command not executable: %s", command.GetName()))
			commandSpan.End()
		}

		if chCtx.HasErrors() {
			commandSpan.SetStatus(codes.Error, "error after execute")
		} else {
			commandSpan.SetStatus(codes.Ok, command.GetName())
		}

		commandSpan.End()

		// Flipflop input/output
		chCtx.Remove(CtxIn)
		chCtx.Add(CtxIn, chCtx.Get(CtxOut))
		chCtx.Remove(CtxOut)
	}

	if !chCtx.HasErrors() {
		chainSpan.SetStatus(codes.Ok, c.GetName())
	} else {
		chainSpan.SetStatus(codes.Error, "chain failed to execute")
	}
	chainSpan.End()
}

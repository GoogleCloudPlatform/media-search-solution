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

package workflow_test

import (
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/workflow"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
)

func TestMediaEmbeddings(t *testing.T) {
	traceCtx, span := tracer.Start(ctx, "generate_embeddings")
	defer span.End()

	chainCtx := cor.NewBaseContext()
	chainCtx.SetContext(traceCtx)

	embeddingWorkflow := workflow.NewMediaEmbeddingGeneratorWorkflow(config, cloudClients)
	embeddingWorkflow.Execute(chainCtx)

	for _, e := range chainCtx.GetErrors() {
		fmt.Printf("Error: %v \n", e)
	}

	assert.False(t, chainCtx.HasErrors())
	span.SetStatus(codes.Ok, "success")
	assert.Nil(t, err)
}

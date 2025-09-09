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
//         kingman (Charlie Wang)

package workflow_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/workflow"
	"github.com/GoogleCloudPlatform/media-search-solution/test"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
)

func TestMediaChain(t *testing.T) {
	traceCtx, span := tracer.Start(ctx, "media-ingestion-test")
	defer span.End()

	mediaIngestion := workflow.NewMediaReaderPipeline(config, cloudClients, "creative-flash", "bin/ffprobe", templateService)

	chainCtx := cor.NewBaseContext()
	chainCtx.SetContext(traceCtx)
	chainCtx.Add(cor.CtxIn, test.GetTestLowResMessageText())

	mediaIngestion.Execute(chainCtx)

	for k, err := range chainCtx.GetErrors() {
		fmt.Printf("Error: (%s): %v\n", k, err)
	}

	if chainCtx.HasErrors() {
		span.SetStatus(codes.Error, "failed to execute media ingestion test")
	}

	assert.False(t, chainCtx.HasErrors())

	span.SetStatus(codes.Ok, "passed - media ingestion test")

	log.Println(chainCtx.Get("MEDIA"))
}

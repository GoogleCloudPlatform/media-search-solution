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
	"context"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/telemetry"
	"github.com/GoogleCloudPlatform/media-search-solution/test"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

var err error
var cloudClients *cloud.ServiceClients
var ctx context.Context
var config *cloud.Config
var templateService *cloud.TemplateService

const tName = "cloud.google.com/media/tests/workflow"

var (
	tracer = otel.Tracer(tName)
	logger = otelslog.NewLogger(tName)
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	// This deferral will automatically close the client that was build from
	// the same context
	defer cancel()

	// Get the config file
	config = test.GetConfig()
	templateService = cloud.NewTemplateService(config)

	telemetry.SetupLogging()
	shutdown, err := telemetry.SetupOpenTelemetry(ctx, config)
	if err != nil {
		panic(err)
	}

	cloudClients, err = cloud.NewCloudServiceClients(ctx, config)
	if err != nil {
		panic(err)
	}
	defer cloudClients.Close()

	logger.Info("completed test setup")

	exitCode := m.Run()
	err = shutdown(ctx)
	if err != nil {
		return
	}
	os.Exit(exitCode)
}

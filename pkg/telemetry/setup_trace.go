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

package telemetry

import (
	"context"
	"errors"
	"log"
	"log/slog"

	"go.opentelemetry.io/otel/sdk/metric"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	telemetryexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semaphoreconversion "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func SetupOpenTelemetry(ctx context.Context, config *cloud.Config) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown combines shutdown functions from multiple OpenTelemetry
	// components into a single function.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Identify your application using resource detection
	res, err := resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semaphoreconversion.ServiceNameKey.String(config.Application.Name),
		),
	)
	if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
		slog.Warn("partial resource", "error", err)
	} else if err != nil {
		slog.Error("resource.New", "error", err)
	}

	// Configure Context Propagation to use the default W3C traceparent format
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	traceExporter, err := telemetryexporter.New(telemetryexporter.WithProjectID(config.Application.GoogleProjectId))
	if err != nil {
		slog.Error("unable to set up tracing", "error", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)

	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)

	mExporter, err := mexporter.New(
		mexporter.WithProjectID(config.Application.GoogleProjectId),
	)

	if err != nil {
		log.Printf("Failed to create exporter: %v", err)
		return nil, err
	}

	mProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(mExporter)),
	)

	// Setup Namespace Meter
	otel.Meter("github.com/GoogleCloudPlatform/media-search-solution")

	shutdownFuncs = append(shutdownFuncs, mProvider.Shutdown)
	otel.SetMeterProvider(mProvider)

	return shutdown, nil
}

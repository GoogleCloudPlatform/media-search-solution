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
	"io"
	"log"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

type spanContextLogHandler struct {
	slog.Handler
}

func handlerWithSpanContext(handler slog.Handler) *spanContextLogHandler {
	return &spanContextLogHandler{Handler: handler}
}

func (t *spanContextLogHandler) Handle(ctx context.Context, record slog.Record) error {
	// Get the SpanContext from the golang Context.
	if s := trace.SpanContextFromContext(ctx); s.IsValid() {
		// Add trace context attributes following Cloud Logging structured log format described
		// in https://cloud.google.com/logging/docs/structured-logging#special-payload-fields

		record.AddAttrs(
			slog.Any("logging.googleapis.com/trace", s.TraceID()),
		)
		record.AddAttrs(
			slog.Any("logging.googleapis.com/spanId", s.SpanID()),
		)
		record.AddAttrs(
			slog.Bool("logging.googleapis.com/trace_sampled", s.TraceFlags().IsSampled()),
		)
	}
	return t.Handler.Handle(ctx, record)
}

func replacer(_ []string, a slog.Attr) slog.Attr {
	// Rename attribute keys to match Cloud Logging structured log format
	switch a.Key {
	case slog.LevelKey:
		a.Key = "severity"
		// Map slog.Level string values to Cloud Logging LogSeverity
		// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
		if level := a.Value.Any().(slog.Level); level == slog.LevelWarn {
			a.Value = slog.StringValue("WARNING")
		}
	case slog.TimeKey:
		a.Key = "timestamp"
	case slog.MessageKey:
		a.Key = "message"
	}
	return a
}

func SetupLogging() {
	// Create a multi writer
	file, _ := os.Create("app.log")
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Set output to standard output (instead of the default standard error)
	log.SetOutput(multiWriter)

	// Set a prefix for log messages
	log.SetPrefix("[INFO] ")

	// Add flags for date and time
	log.SetFlags(log.Ldate | log.Ltime)

	// Setup the SLOG behavior
	// Use json as our base logging format.
	jsonHandler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{ReplaceAttr: replacer})

	// Add span context attributes when Context is passed to logging calls.
	instrumentedHandler := handlerWithSpanContext(jsonHandler)

	// Set this handler as the global slog handler.
	slog.SetDefault(slog.New(instrumentedHandler))
	slog.SetLogLoggerLevel(slog.LevelInfo)
}

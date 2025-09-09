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
//         jaycherian (Jay Cherian)
//         kingman (Charlie Wang)

package cloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel/metric"

	"github.com/BurntSushi/toml"
	"google.golang.org/genai"
)

// Cloud Constants
const (
	ConfigFileBaseName  = ".env"
	ConfigFileExtension = ".toml"
	ConfigSeparator     = "."
	EnvConfigFilePrefix = "GCP_CONFIG_PREFIX"
	EnvConfigRuntime    = "GCP_RUNTIME"
	MaxRetries          = 3
)

// Simple utility to see if a file exists
func fileExists(in string) bool {
	_, err := os.Stat(in)
	return !errors.Is(err, os.ErrNotExist)
}

// LoadConfig The configuration loader, a hierarchical loader that allows environment overrides.
func LoadConfig(baseConfig interface{}) {
	configurationFilePrefix := os.Getenv(EnvConfigFilePrefix)
	if len(configurationFilePrefix) > 0 && !strings.HasSuffix(configurationFilePrefix, string(os.PathSeparator)) {
		configurationFilePrefix = configurationFilePrefix + string(os.PathSeparator)
	}

	runtimeEnvironment := os.Getenv(EnvConfigRuntime)
	if runtimeEnvironment == "" {
		runtimeEnvironment = "test"
	}

	// Read Base Config
	baseConfigFileName := configurationFilePrefix + ConfigFileBaseName + ConfigFileExtension
	fmt.Printf("Base Configuration File: %s\n", baseConfigFileName)

	// Override with environment config
	envConfigFileName := configurationFilePrefix + ConfigFileBaseName + ConfigSeparator + runtimeEnvironment + ConfigFileExtension
	fmt.Printf("Environment Configuration File: %s\n", envConfigFileName)

	if fileExists(baseConfigFileName) {
		_, err := toml.DecodeFile(baseConfigFileName, baseConfig)
		if err != nil {
			log.Fatalf("failed to decode base configuration file %s with error: %s", baseConfigFileName, err)
		}
	}

	if fileExists(envConfigFileName) {
		_, err := toml.DecodeFile(envConfigFileName, baseConfig)
		if err != nil {
			log.Fatalf("failed to decode environment configuration file: %s with error: %s", envConfigFileName, err)
		}
	}
}

// GenerateMultiModalResponse A GenAI helper function for executing multi-modal requests with a retry limit.
func GenerateMultiModalResponse(
	ctx context.Context,
	inputTokenCounter metric.Int64Counter,
	outputTokenCounter metric.Int64Counter,
	retryCounter metric.Int64Counter,
	tryCount int,
	model *QuotaAwareGenerativeAIModel,
	systemInstruction string,
	contents []*genai.Content,
	outputSchema *genai.Schema) (value string, err error) {
	resp, err := model.GenerateContent(ctx, systemInstruction, contents, outputSchema)
	inputTokenCounter.Add(ctx, int64(resp.UsageMetadata.PromptTokenCount))
	outputTokenCounter.Add(ctx, int64(resp.UsageMetadata.CandidatesTokenCount))
	if err != nil {
		if tryCount < MaxRetries {
			retryCounter.Add(ctx, 1)
			return GenerateMultiModalResponse(ctx, inputTokenCounter, outputTokenCounter, retryCounter, tryCount+1, model, systemInstruction, contents, outputSchema)
		} else {
			return "", err
		}
	}
	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				value += fmt.Sprint(part.Text)
			}
		}
	}
	if len(value) == 0 {
		log.Println("Empty response from model, retrying...")
		if tryCount < MaxRetries {
			retryCounter.Add(ctx, 1)
			return GenerateMultiModalResponse(ctx, inputTokenCounter, outputTokenCounter, retryCounter, tryCount+1, model, systemInstruction, contents, outputSchema)
		} else {
			return "", errors.New("no candidates returned from model after retries")
		}
	}
	return value, nil
}

// NewTextPart A delegate method for creating text parts
func NewTextPart(in string) []*genai.Content {
	return genai.Text(in)
}

// NewFileData A delegate method for creating File Data parts.
func NewFileData(in string, mimeType string) genai.FileData {
	return genai.FileData{FileURI: in, MIMEType: mimeType}
}

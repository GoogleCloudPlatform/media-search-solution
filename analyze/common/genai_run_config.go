// Copyright 2025 Google, LLC
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
// Author: kingman (Charlie Wang)

package common

import (
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/genai"
)

const (
	GENAI_INPUT_FILE_TYPE = "video/mp4"
)

type GenaiRunConfig struct {
	BasicRunConfig
	CloudConfig        *cloud.Config
	AgentModels        map[string]*cloud.QuotaAwareGenerativeAIModel
	TemplateService    *cloud.TemplateService
	Meter              metric.Meter
	GenAIClient        *genai.Client
	GenAIContentCaches map[string]*genai.CachedContent
	BigQueryClient     *bigquery.Client
	GenAIEmbedding     *genai.Models
}

func NewGenaiRunConfig() (*GenaiRunConfig, error) {
	basicRunConfig, err := NewBasicRunConfig()
	if err != nil {
		return nil, err
	}
	cloudConfig, err := loadCloudConfig()
	if err != nil {
		return nil, err
	}
	cloudClients, err := cloud.NewCloudServiceClients(basicRunConfig.Ctx, cloudConfig)
	if err != nil {
		return nil, err
	}
	templateService := cloud.NewTemplateService(cloudConfig)

	meter := otel.Meter("github.com/GoogleCloudPlatform/media-search-solution")
	config := &GenaiRunConfig{
		BasicRunConfig:  *basicRunConfig,
		CloudConfig:     cloudConfig,
		AgentModels:     cloudClients.AgentModels,
		TemplateService: templateService,
		Meter:           meter,
		GenAIClient:     cloudClients.GenAIClient,
		BigQueryClient:  cloudClients.BiqQueryClient,
		GenAIEmbedding:  cloudClients.EmbeddingModels["multi-lingual"],
	}
	config.SetStorageClient(cloudClients.StorageClient)
	return config, nil

}

func loadCloudConfig() (_ *cloud.Config, err error) {
	GCP_CONFIG_PREFIX := os.Getenv(cloud.EnvConfigFilePrefix)
	if len(GCP_CONFIG_PREFIX) == 0 {
		return nil, fmt.Errorf("required environment variable %s is not set", cloud.EnvConfigFilePrefix)
	}

	configRuntimeValue := os.Getenv(cloud.EnvConfigRuntime)
	if configRuntimeValue == "" {
		err = os.Setenv(cloud.EnvConfigRuntime, "local")
		if err != nil {
			return nil, err
		}
	}

	config := cloud.NewConfig()
	cloud.LoadConfig(&config)

	return config, err
}

func (config *GenaiRunConfig) GetInputFileGCSURI() string {
	return "gs://" + config.InputBucket + "/" + config.InputFile
}

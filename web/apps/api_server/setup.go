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

package main

import (
	"context"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/services"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/workflow"
)

type StateManager struct {
	config        *cloud.Config
	cloud         *cloud.ServiceClients
	searchService *services.SearchService
	mediaService  *services.MediaService
}

var state = &StateManager{}

func SetupOS() (err error) {
	configPrefixValue := os.Getenv(cloud.EnvConfigFilePrefix)
	log.Printf("Config file prefix: %s\n", configPrefixValue)
	if configPrefixValue == "" {
		err = os.Setenv(cloud.EnvConfigFilePrefix, "configs")
		if err != nil {
			return err
		}
	}

	configRuntimeValue := os.Getenv(cloud.EnvConfigRuntime)
	if configRuntimeValue == "" {
		err = os.Setenv(cloud.EnvConfigRuntime, "local")
		if err != nil {
			return err
		}
	}
	return err
}

func GetConfig() *cloud.Config {
	if state.config == nil {
		err := SetupOS()
		if err != nil {
			log.Fatalf("failed to setup os for testing: %v\n", err)
		}
		// Create a default cloud config
		config := cloud.NewConfig()
		// Load it from the TOML files
		cloud.LoadConfig(&config)
		state.config = config
	}
	return state.config
}

func InitState(ctx context.Context) {
	// Get the config file
	config := GetConfig()

	cloudClients, err := cloud.NewCloudServiceClients(ctx, config)
	if err != nil {
		panic(err)
	}

	state.cloud = cloudClients

	datasetName := config.BigQueryDataSource.DatasetName
	mediaTableName := config.BigQueryDataSource.MediaTable
	embeddingTableName := config.BigQueryDataSource.EmbeddingTable

	state.searchService = &services.SearchService{
		BigqueryClient: cloudClients.BiqQueryClient,
		EmbeddingModel: cloudClients.EmbeddingModels["multi-lingual"],
		DatasetName:    datasetName,
		MediaTable:     mediaTableName,
		EmbeddingTable: embeddingTableName,
		ModelName:      config.EmbeddingModels["multi-lingual"].Model,
	}

	state.mediaService = &services.MediaService{
		BigqueryClient: cloudClients.BiqQueryClient,
		DatasetName:    datasetName,
		MediaTable:     mediaTableName,
	}

	embeddingGenerator := workflow.NewMediaEmbeddingGeneratorWorkflow(config, cloudClients)
	embeddingGenerator.StartTimer()

	SetupListeners(config, cloudClients, cloud.NewTemplateService(config), ctx)

}

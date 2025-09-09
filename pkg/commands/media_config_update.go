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

package commands

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
)

type MediaConfigUpdateCommand struct {
	cor.BaseCommand
	config          *cloud.Config
	templateService *cloud.TemplateService
}

func NewMediaConfigUpdateCommand(name string, config *cloud.Config, templateService *cloud.TemplateService) *MediaConfigUpdateCommand {
	return &MediaConfigUpdateCommand{
		BaseCommand:     *cor.NewBaseCommand(name),
		config:          config,
		templateService: templateService}
}

func (m *MediaConfigUpdateCommand) Execute(context cor.Context) {
	gcsFile := context.Get(cloud.GetGCSObjectName()).(*cloud.GCSObject)
	configurationFilePrefix := os.Getenv(cloud.EnvConfigFilePrefix)
	if len(configurationFilePrefix) > 0 && !strings.HasSuffix(configurationFilePrefix, string(os.PathSeparator)) {
		configurationFilePrefix = configurationFilePrefix + string(os.PathSeparator)
	}

	localConfigFile := configurationFilePrefix + gcsFile.Name

	m.WaitForTheLocalFileToUpdate(localConfigFile)

	newConfig := cloud.NewConfig()
	// Load the configuration values for the updated config files
	cloud.LoadConfig(&newConfig)
	// Replace the current config with the new one
	m.config.Replace(newConfig)
	// Update the templates with the new config values
	m.templateService.UpdateTemplates()

	m.GetSuccessCounter().Add(context.GetContext(), 1)
}

func (m *MediaConfigUpdateCommand) WaitForTheLocalFileToUpdate(localFile string) {
	// it can take some time to sync the file from the bucket to the local filesystem.
	// We check for the file's existence and modification time to ensure we have the latest version.
	const recentThreshold = 30 * time.Second
	for i := range FileCheckRetries {
		fileInfo, err := os.Stat(localFile)
		if err == nil {
			if time.Since(fileInfo.ModTime()) < recentThreshold {
				log.Printf("Configuration file %s has been updated recently.", localFile)
				return
			}
		}
		log.Printf("waiting for configuration file to be updated: %s, attempt %d/%d", localFile, i+1, FileCheckRetries)
		time.Sleep(FileCheckDelay)
	}
	log.Printf("Configuration file %s not updated after several retries. Proceeding with existing config.", localFile)
}

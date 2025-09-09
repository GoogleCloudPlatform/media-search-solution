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

package workflow

import (
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/commands"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
)

type MediaConfigUpdateWorkflow struct {
	cor.BaseCommand
	chain           cor.Chain
	config          *cloud.Config
	templateService *cloud.TemplateService
}

func (m *MediaConfigUpdateWorkflow) Execute(context cor.Context) {
	m.chain.Execute(context)
}
func (m *MediaConfigUpdateWorkflow) initializeChain() {
	out := cor.NewBaseChain(m.GetName())

	out.AddCommand(commands.NewMediaTriggerToGCSObject("gcs-topic-listener"))

	out.AddCommand(commands.NewMediaConfigUpdateCommand("config-update-command", m.config, m.templateService))

	m.chain = out
}

func NewMediaConfigUpdateWorkflow(config *cloud.Config, templateService *cloud.TemplateService) *MediaConfigUpdateWorkflow {
	out := &MediaConfigUpdateWorkflow{
		BaseCommand:     *cor.NewBaseCommand("media-config-update-workflow"),
		config:          config,
		templateService: templateService}
	out.initializeChain()
	return out
}

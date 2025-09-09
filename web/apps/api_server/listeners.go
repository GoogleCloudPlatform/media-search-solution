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

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/workflow"
)

func SetupListeners(config *cloud.Config, cloudClients *cloud.ServiceClients, templateService *cloud.TemplateService, ctx context.Context) {
	// TODO - Externalize the destination topic and ffmpeg command
	mediaResizeWorkflow := workflow.NewMediaResizeWorkflow(config, cloudClients, "bin/ffmpeg", &model.MediaFormatFilter{Width: "240"})
	cloudClients.PubSubListeners["HiResTopic"].SetCommand(mediaResizeWorkflow)
	cloudClients.PubSubListeners["HiResTopic"].Listen(ctx)

	mediaIngestion := workflow.NewMediaReaderPipeline(config, cloudClients, "creative-flash", "bin/ffprobe", templateService)

	cloudClients.PubSubListeners["LowResTopic"].SetCommand(mediaIngestion)
	cloudClients.PubSubListeners["LowResTopic"].Listen(ctx)

	mediaConfigUpdateWorkflow := workflow.NewMediaConfigUpdateWorkflow(config, templateService)
	cloudClients.PubSubListeners["ConfigTopic"].SetCommand(mediaConfigUpdateWorkflow)
	cloudClients.PubSubListeners["ConfigTopic"].Listen(ctx)
}

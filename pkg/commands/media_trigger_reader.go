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

package commands

import (
	"encoding/json"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
)

type MediaTriggerToGCSObject struct {
	cor.BaseCommand
}

func NewMediaTriggerToGCSObject(name string) *MediaTriggerToGCSObject {
	return &MediaTriggerToGCSObject{BaseCommand: *cor.NewBaseCommand(name)}
}

func (c *MediaTriggerToGCSObject) Execute(context cor.Context) {
	in := context.Get(c.GetInputParam()).(string)
	var out cloud.GCSPubSubNotification
	err := json.Unmarshal([]byte(in), &out)
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), err)
		return
	}

	c.GetSuccessCounter().Add(context.GetContext(), 1)

	msg := &cloud.GCSObject{Bucket: out.Bucket, Name: out.Name, MIMEType: out.ContentType}
	context.Add(cloud.GetGCSObjectName(), msg)
	context.Add(c.GetOutputParam(), msg)
}

// Package commands Copyright 2024 Google, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: rrmcguinness (Ryan McGuinness)
//         kingman (Charlie Wang)

package commands

import (
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
)

type MediaSummaryJsonToStruct struct {
	cor.BaseCommand
}

func NewMediaSummaryJsonToStruct(name string, outputParamName string) *MediaSummaryJsonToStruct {
	out := MediaSummaryJsonToStruct{BaseCommand: *cor.NewBaseCommand(name)}
	out.OutputParamName = outputParamName
	return &out
}

func (s *MediaSummaryJsonToStruct) Execute(context cor.Context) {
	in := context.Get(s.GetInputParam()).(string)
	gcsFile := context.Get(cloud.GetGCSObjectName()).(*cloud.GCSObject)

	doc := &model.MediaSummary{}
	err := json.Unmarshal([]byte(in), &doc)
	if err != nil {
		s.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(s.GetName(), err)
		return
	}
	s.GetSuccessCounter().Add(context.GetContext(), 1)
	doc.MediaUrl = fmt.Sprintf("https://storage.mtls.cloud.google.com/%s/%s", gcsFile.Bucket, gcsFile.Name)
	context.Add(s.GetOutputParam(), doc)
	context.Add(cor.CtxOut, doc)
}

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
	"context"
)

const (
	StepCompleted               = "completed"
	GENERATE_PROXY_STEP         = "ims_generate_proxy"
	CONTENT_LENGTH_STEP         = "ims_content_length"
	CONTENT_TYPE_STEP           = "ims_content_type"
	CONTENT_SUMMARY_STEP        = "ims_content_summary"
	SEGMENT_SUMMARY_STEP_PREFIX = "ims_segment_summary_"
	PERSIST_STEP                = "ims_persist"
	EMBEDDING_STEP              = "ims_generate_embeddings"
)

type RunConfig struct {
	InputFile       string
	InputBucket     string
	StepMetadataKey string
	MountPoint      string
	Ctx             context.Context
}

type StepStatus struct {
	Output string `json:"output"`
	Status string `json:"status"`
}

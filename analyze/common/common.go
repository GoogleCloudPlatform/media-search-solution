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

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
	"text/template"

	"google.golang.org/genai"
)

// DefaultSafetySettings Default System Settings for GenAI agents
var DefaultSafetySettings = []*genai.SafetySetting{
	{
		Category:  genai.HarmCategoryDangerousContent,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
	{
		Category:  genai.HarmCategoryHarassment,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
	{
		Category:  genai.HarmCategoryHateSpeech,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
	{
		Category:  genai.HarmCategorySexuallyExplicit,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
}

// BigQueryDataSource represents the configuration for a BigQuery data source.
type BigQueryDataSource struct {
	DatasetName    string `toml:"dataset"`         // The name of the BigQuery dataset.
	MediaTable     string `toml:"media_table"`     // The name of the BigQuery table containing media information.
	EmbeddingTable string `toml:"embedding_table"` // The name of the BigQuery table containing embedding vectors.
}

// PromptTemplates holds the templates for different types of prompts.
type PromptTemplates struct {
	SystemInstructions string `toml:"system_instructions"` // The system instructions for the LLM.
	SummaryPrompt      string `toml:"summary"`             // The template for generating summaries.
	SegmentPrompt      string `toml:"segment"`             // The template for generating segment descriptions.
}

// PromptTemplate holds the templates for generating summaries and segments.
type PromptTemplate struct {
	SystemInstructions string
	SummaryPrompt      *template.Template
	SegmentPrompt      *template.Template
}

// VertexAiEmbeddingModel represents the configuration for a Vertex AI embedding model.
type VertexAiEmbeddingModel struct {
	Model                string `toml:"model"`                   // The name of the Vertex AI embedding model.
	MaxRequestsPerMinute int    `toml:"max_requests_per_minute"` // The maximum number of requests allowed per minute.
}

// VertexAiLLMModel represents the configuration for a Vertex AI large language model (LLM).
type VertexAiLLMModel struct {
	Model              string  `toml:"model"`               // The name of the Vertex AI LLM.
	SystemInstructions string  `toml:"system_instructions"` // The system instructions for the LLM.
	Temperature        float32 `toml:"temperature"`         // The temperature parameter for the LLM.
	TopP               float32 `toml:"top_p"`               // The top_p parameter for the LLM.
	TopK               float32 `toml:"top_k"`               // The top_k parameter for the LLM.
	MaxTokens          int32   `toml:"max_tokens"`          // The maximum number of tokens for the LLM output.
	OutputFormat       string  `toml:"output_format"`       // The desired output format for the LLM.
	EnableGoogle       bool    `toml:"enable_google"`       // Whether to enable Google Search for the LLM.
	RateLimit          int     `toml:"rate_limit"`          // The rate limit for the LLM in requests per second.
}

// TopicSubscription represents the configuration for a Pub/Sub topic subscription.
type TopicSubscription struct {
	Name             string `toml:"name"`               // The name of the Pub/Sub subscription.
	DeadLetterTopic  string `toml:"dead_letter_topic"`  // The name of the dead-letter topic for the subscription.
	TimeoutInSeconds int    `toml:"timeout_in_seconds"` // The timeout for the subscription in seconds.
}

// Storage represents the configuration for storage buckets.
type Storage struct {
	HiResInputBucket   string `toml:"high_res_input_bucket"` // The name of the bucket for high-resolution input files.
	LowResOutputBucket string `toml:"low_res_output_bucket"` // The name of the bucket for low-resolution output files.
	GCSFuseMountPoint  string `toml:"gcs_fuse_mount_point"`  // The mount point for GCS FUSE.
}

type Category struct {
	Name               string `toml:"name"`
	Definition         string `toml:"definition"`
	SystemInstructions string `toml:"system_instructions"`
	Summary            string `toml:"summary"`
	Segment            string `toml:"segment"`
}

type ContentType struct {
	Types          []string `toml:"types"`           // A list of content types.
	PromptTemplate string   `toml:"prompt_template"` // The template for generating content type
	DefaultType    string   `toml:"default_type"`    // The default content type to use if none is matched.
}

// Config represents the overall configuration for the application.
type Config struct {
	Application struct {
		Name            string `toml:"name"`              // The name of the application.
		GoogleProjectId string `toml:"google_project_id"` // The Google Cloud project ID.
		GoogleLocation  string `toml:"location"`          // The Google Cloud location.
		ThreadPoolSize  int    `toml:"thread_pool_size"`  // The size of the thread pool.
	} `toml:"application"`
	Storage            Storage                           `toml:"storage"`               // Storage configuration.
	BigQueryDataSource BigQueryDataSource                `toml:"big_query_data_source"` // BigQuery data source configuration.
	PromptTemplates    map[string]PromptTemplates        `toml:"prompt_templates"`      // Prompt templates configuration.
	TopicSubscriptions map[string]TopicSubscription      `toml:"topic_subscriptions"`   // Pub/Sub topic subscriptions configuration.
	EmbeddingModels    map[string]VertexAiEmbeddingModel `toml:"embedding_models"`      // Vertex AI embedding models configuration.
	AgentModels        map[string]VertexAiLLMModel       `toml:"agent_models"`          // Vertex AI LLM models configuration.
	Categories         map[string]Category               `toml:"categories"`            // A list of category definitions and LLM overrides.
	ContentType        ContentType                       `toml:"content_type"`          // Content type configuration.
}

func (c *Config) Replace(newConfig *Config) {
	c.Application = newConfig.Application
	c.Storage = newConfig.Storage
	c.BigQueryDataSource = newConfig.BigQueryDataSource
	c.PromptTemplates = newConfig.PromptTemplates
	c.TopicSubscriptions = newConfig.TopicSubscriptions
	c.EmbeddingModels = newConfig.EmbeddingModels
	c.AgentModels = newConfig.AgentModels
	c.Categories = newConfig.Categories
	c.ContentType = newConfig.ContentType
}

// NewConfig creates a new Config instance with initialized maps.
func NewConfig() *Config {
	return &Config{
		TopicSubscriptions: make(map[string]TopicSubscription),
		EmbeddingModels:    make(map[string]VertexAiEmbeddingModel),
		AgentModels:        make(map[string]VertexAiLLMModel),
		Categories:         make(map[string]Category),
	}
}

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
// Author: rrmcguinness (Ryan McGuinness)
//         jaycherian (Jay Cherian)
//         kingman (Charlie Wang)

package cloud

import (
	"context"
	"errors"
	"log"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/genai"
)

// QuotaAwareGenerativeAIModel wraps a genai.GenerativeModel with rate limiting.
type QuotaAwareGenerativeAIModel struct {
	GenerativeContentConfig *genai.GenerateContentConfig // The configuration for LLM content genration.
	ModelName               string
	ModelHandle             *genai.Models
	RateLimit               rate.Limiter // The rate limiter for the LLM.
}

// NewQuotaAwareModel creates a new QuotaAwareGenerativeAIModel with the given rate limit.
func NewQuotaAwareModel(wrapped *genai.GenerateContentConfig, modelName string, modelHandle *genai.Models, requestsPerSecond int) *QuotaAwareGenerativeAIModel {
	return &QuotaAwareGenerativeAIModel{
		GenerativeContentConfig: wrapped,
		ModelName:               modelName,
		ModelHandle:             modelHandle,
		RateLimit:               *rate.NewLimiter(rate.Every(time.Second/1), requestsPerSecond),
	}
}

// GenerateContent generates content using the wrapped LLM with rate limiting.
func (q *QuotaAwareGenerativeAIModel) GenerateContent(ctx context.Context, systemInstruction string, contents []*genai.Content, outputSchema *genai.Schema) (resp *genai.GenerateContentResponse, err error) {
	// Create a copy of the generative content config to avoid modifying the original.
	config := *q.GenerativeContentConfig

	// set the desired output schema, take it from
	if outputSchema != nil {
		config.ResponseSchema = outputSchema
	}

	if systemInstruction != "" {
		config.SystemInstruction = genai.NewContentFromText(systemInstruction, genai.RoleUser)
	}
	// Check if the rate limit allows a request.
	if q.RateLimit.Allow() {
		// If allowed, make the request to the LLM.
		resp, err = q.ModelHandle.GenerateContent(ctx, q.ModelName, contents, &config)
		if err != nil {
			log.Printf("Error generating content: %v", err)
			// If there's an error, check the retry count from the context.
			retryCount, ok := ctx.Value("retry").(int)
			if !ok {
				// This is the first attempt.
				retryCount = 0
			}
			if retryCount > 3 {
				// If retry count exceeds the limit, return an error.
				return nil, errors.New("failed generation on max retries")
			}
			// If retries are allowed, wait for one minute and try again.
			errCtx := context.WithValue(ctx, "retry", retryCount+1)
			time.Sleep(time.Minute * 1)
			return q.ModelHandle.GenerateContent(errCtx, q.ModelName, contents, &config)
		}
		// If successful, return the response.
		return resp, err
	} else {
		// If rate limit is exceeded, wait for 5 seconds and try again.
		time.Sleep(time.Second * 5)
		return q.GenerateContent(ctx, systemInstruction, contents, outputSchema)
	}
}

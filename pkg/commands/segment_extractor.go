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
// Author: rrmcguinness (Ryan McGuinness)
//         jaycherian (Jay Cherian)
//         kingman (Charlie Wang)

package commands

import (
	"bytes"
	goctx "context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"go.opentelemetry.io/otel/metric"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genai"
)

type SegmentExtractor struct {
	cor.BaseCommand
	generativeAIModel        *cloud.QuotaAwareGenerativeAIModel
	templateService          *cloud.TemplateService
	numberOfWorkers          int
	geminiInputTokenCounter  metric.Int64Counter
	geminiOutputTokenCounter metric.Int64Counter
	geminiRetryCounter       metric.Int64Counter
	contentTypeParamName     string
}

func NewSegmentExtractor(
	name string,
	model *cloud.QuotaAwareGenerativeAIModel,
	templateService *cloud.TemplateService,
	numberOfWorkers int,
	contentTypeParamName string) *SegmentExtractor {
	out := &SegmentExtractor{
		BaseCommand:          *cor.NewBaseCommand(name),
		generativeAIModel:    model,
		templateService:      templateService,
		numberOfWorkers:      numberOfWorkers,
		contentTypeParamName: contentTypeParamName}

	out.geminiInputTokenCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.input", out.GetName()))
	out.geminiOutputTokenCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.ouput", out.GetName()))
	out.geminiRetryCounter, _ = out.GetMeter().Int64Counter(fmt.Sprintf("%s.gemini.token.retry", out.GetName()))

	return out
}

func (s *SegmentExtractor) IsExecutable(context cor.Context) bool {
	return context != nil &&
		context.Get(s.GetInputParam()) != nil &&
		context.Get(cloud.GetGCSObjectName()) != nil
}

func (s *SegmentExtractor) Execute(context cor.Context) {
	summary := context.Get(s.GetInputParam()).(*model.MediaSummary)
	gcsFile := context.Get(cloud.GetGCSObjectName()).(*cloud.GCSObject)
	gcsFileLink := fmt.Sprintf("gs://%s/%s", gcsFile.Bucket, gcsFile.Name)
	mediaType := context.Get(s.contentTypeParamName).(string)
	videoFile := &genai.FileData{
		FileURI:  gcsFileLink,
		MIMEType: gcsFile.MIMEType,
	}

	exampleSegment := model.GetExampleSegment()
	exampleJson, _ := json.Marshal(exampleSegment)
	exampleText := string(exampleJson)

	// Create a human-readable cast
	castString := ""
	for _, cast := range summary.Cast {
		castString += fmt.Sprintf("%s - %s\n", cast.CharacterName, cast.ActorName)
	}
	summaryText := fmt.Sprintf("Title:%s\nSummary:\n\n%s\nCast:\n\n%v\n", summary.Title, summary.Summary, castString)

	var wg sync.WaitGroup
	jobs := make(chan *SegmentJob, len(summary.SegmentTimeStamps))
	results := make(chan *SegmentResponse, len(summary.SegmentTimeStamps))

	// Create worker pool
	for w := 1; w <= s.numberOfWorkers; w++ {
		wg.Add(1)
		go segmentWorker(jobs, results, &wg)
	}

	// Execute all segments against the worker pool
	for i, ts := range summary.SegmentTimeStamps {
		job := CreateJob(context.GetContext(), s.Tracer, s.geminiInputTokenCounter, s.geminiOutputTokenCounter, s.geminiRetryCounter, i, s.GetName(), summaryText, exampleText, *s.templateService.GetTemplateBy(mediaType).SegmentPrompt, videoFile, s.generativeAIModel, ts)
		jobs <- job
	}

	close(jobs)
	wg.Wait()
	close(results)

	// Aggregate the responses
	segmentData := make([]string, 0)
	for r := range results {
		if r.err != nil {
			s.GetErrorCounter().Add(context.GetContext(), 1)
			context.AddError(s.GetName(), r.err)
		} else {

			segmentData = append(segmentData, r.value)
		}
	}

	if !context.HasErrors() {
		s.GetSuccessCounter().Add(context.GetContext(), 1)
	}

	context.Add(s.GetOutputParam(), segmentData)
	context.Add(cor.CtxOut, segmentData)
}

type SegmentResponse struct {
	value string
	err   error
}

type SegmentJob struct {
	workerId                 int
	ctx                      goctx.Context
	geminiInputTokenCounter  metric.Int64Counter
	geminiOutputTokenCounter metric.Int64Counter
	geminiRetryCounter       metric.Int64Counter
	timeSpan                 *model.TimeSpan
	span                     trace.Span
	contents                 []*genai.Content
	model                    *cloud.QuotaAwareGenerativeAIModel
	err                      error
}

func (s *SegmentJob) Close(status codes.Code, description string) {
	s.span.SetStatus(status, description)
	s.span.End()
}

func CreateJob(
	ctx goctx.Context,
	tracer trace.Tracer,
	geminiInputTokenCounter metric.Int64Counter,
	geminiOutputTokenCounter metric.Int64Counter,
	geminiRetryCounter metric.Int64Counter,
	workerId int,
	commandName string,
	summaryText string,
	exampleText string,
	template template.Template,
	videoFile *genai.FileData,
	model *cloud.QuotaAwareGenerativeAIModel,
	timeSpan *model.TimeSpan,
) *SegmentJob {
	segmentCtx, segmentSpan := tracer.Start(ctx, fmt.Sprintf("%s_genai", commandName))
	segmentSpan.SetAttributes(
		attribute.Int("sequence", workerId),
		attribute.String("start", timeSpan.Start),
		attribute.String("end", timeSpan.End),
	)
	vocabulary := make(map[string]string)
	vocabulary["SEQUENCE"] = fmt.Sprintf("%d", workerId)
	vocabulary["SUMMARY_DOCUMENT"] = summaryText
	vocabulary["TIME_START"] = timeSpan.Start
	vocabulary["TIME_END"] = timeSpan.End
	vocabulary["EXAMPLE_JSON"] = exampleText

	var doc bytes.Buffer
	err := template.Execute(&doc, vocabulary)
	if err != nil {
		return &SegmentJob{err: err}
	}
	tsPrompt := doc.String()

	contents := []*genai.Content{
		{Parts: []*genai.Part{
			genai.NewPartFromText(tsPrompt),
			genai.NewPartFromURI(videoFile.FileURI, videoFile.MIMEType),
		},
			Role: "user"},
	}

	return &SegmentJob{workerId: workerId,
		ctx:                      segmentCtx,
		geminiInputTokenCounter:  geminiInputTokenCounter,
		geminiOutputTokenCounter: geminiOutputTokenCounter,
		geminiRetryCounter:       geminiRetryCounter,
		timeSpan:                 timeSpan, span: segmentSpan, contents: contents, model: model}
}

// Create a worker function for parallel work streams
func segmentWorker(jobs <-chan *SegmentJob, results chan<- *SegmentResponse, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		if j.err == nil {
			out, err := cloud.GenerateMultiModalResponse(j.ctx, j.geminiInputTokenCounter, j.geminiOutputTokenCounter, j.geminiRetryCounter, 0, j.model, "", j.contents, model.NewSegmentExtractorSchema())
			if err != nil {
				j.Close(codes.Error, "segment extract failed")
				results <- &SegmentResponse{err: err}
				return
			}
			if len(strings.Trim(out, " ")) > 0 && out != "{}" {
				results <- &SegmentResponse{value: out, err: nil}
			}
			j.Close(codes.Ok, "completed segment")
		} else {
			results <- &SegmentResponse{value: "", err: j.err}
		}
	}
}

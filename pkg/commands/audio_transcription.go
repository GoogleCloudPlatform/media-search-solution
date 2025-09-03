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

package commands

import (
	"fmt"
	"log"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/GoogleCloudPlatform/solutions/media/pkg/cloud"
	"github.com/GoogleCloudPlatform/solutions/media/pkg/cor"
	"google.golang.org/api/option"
)

type AudioTranscriptionCommand struct {
	cor.BaseCommand
	config *cloud.Config
}

func NewAudioTranscriptionCommand(name string, config *cloud.Config, outputParamName string) *AudioTranscriptionCommand {
	out := AudioTranscriptionCommand{
		BaseCommand: *cor.NewBaseCommand(name),
		config:      config,
	}
	out.OutputParamName = outputParamName
	return &out
}

func (c *AudioTranscriptionCommand) Execute(context cor.Context) {
	inputFile := context.Get(c.GetInputParam()).(string)

	ctx := context.GetContext()
	client, err := speech.NewClient(ctx, option.WithEndpoint(c.config.Application.GoogleLocation+"-speech.googleapis.com"))
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("failed to create speech client: %w", err))
		return
	}
	defer client.Close()

	config := &speechpb.RecognitionConfig{
		DecodingConfig: &speechpb.RecognitionConfig_AutoDecodingConfig{},
		LanguageCodes:  []string{"en-US"},
		Model:          "chirp_2",
		Features: &speechpb.RecognitionFeatures{
			EnableWordTimeOffsets:      true,
			EnableAutomaticPunctuation: true,
		},
	}

	req := &speechpb.BatchRecognizeRequest{
		Recognizer: fmt.Sprintf("projects/%s/locations/%s/recognizers/_", c.config.Application.GoogleProjectId, c.config.Application.GoogleLocation),
		Config:     config,
		Files: []*speechpb.BatchRecognizeFileMetadata{
			{
				AudioSource: &speechpb.BatchRecognizeFileMetadata_Uri{Uri: inputFile},
			},
		},
		RecognitionOutputConfig: &speechpb.RecognitionOutputConfig{
			Output: &speechpb.RecognitionOutputConfig_GcsOutputConfig{
				GcsOutputConfig: &speechpb.GcsOutputConfig{
					Uri: fmt.Sprintf("gs://%s", c.config.Storage.AudioBucket),
				},
			},
			OutputFormatConfig: &speechpb.OutputFormatConfig{
				Srt: &speechpb.SrtOutputFileFormatConfig{},
			},
		},
	}
	op, err := client.BatchRecognize(ctx, req)
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("error running recognize %w", err))
		return
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("error waiting for recognize %w", err))
		return
	}

	fileResult, ok := resp.Results[inputFile]
	if !ok {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("no result found for file: %s", inputFile))
		return
	}

	if fileResult.Error != nil {
		c.GetErrorCounter().Add(context.GetContext(), 1)
		context.AddError(c.GetName(), fmt.Errorf("error in transcription result for %s: %v", inputFile, fileResult.Error))
		return
	}

	transcriptResultUri := fileResult.GetCloudStorageResult().GetSrtFormatUri()
	log.Printf("Transcription result written to: %s", transcriptResultUri)
	c.GetSuccessCounter().Add(context.GetContext(), 1)
	context.Add(c.GetOutputParam(), transcriptResultUri)
	context.Add(cor.CtxOut, transcriptResultUri)
}

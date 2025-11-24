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
// Author: rrmcguinness (Ryan McGuinness)
//         kingman (Charlie Wang)

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
)

func get_segment_summaries(genaiRunConfig *common.GenaiRunConfig) {
	stepConfig, err := common.NewGenaiStepConfig(common.SEGMENT_SUMMARY_STEP_PREFIX+"all", genaiRunConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	stepConfig.StepLogic = getSegmentSummariesLogicFunc(stepConfig)
	stepConfig.RunStep()
}

func getSegmentSummariesLogicFunc(config *common.GenaiStepConfig) func() (string, error) {
	return func() (string, error) {
		dependentSteps := []string{
			common.CONTENT_TYPE_STEP,
			common.CONTENT_SUMMARY_STEP,
		}
		inputValues := config.BasicRunConfig.GetStepsOutput(dependentSteps)

		contentSummaryObj := &model.MediaSummary{}
		if err := json.Unmarshal([]byte(inputValues[common.CONTENT_SUMMARY_STEP]), &contentSummaryObj); err != nil {
			return "", err
		}

		type result struct {
			stepKey string
			output  string
			err     error
		}

		// 1. Get the status of all segment summary steps in a single GCS read
		allSegmentStepKeys := make([]string, len(contentSummaryObj.SegmentTimeStamps))
		for i := range contentSummaryObj.SegmentTimeStamps {
			allSegmentStepKeys[i] = getSegmentSummaryStepKey(i)
		}
		existingStatuses := config.BasicRunConfig.GetStepsStatus(allSegmentStepKeys)

		// 2. Filter out segments that are already completed
		var segmentsToProcess []int
		for i := range contentSummaryObj.SegmentTimeStamps {
			stepKey := getSegmentSummaryStepKey(i)
			if status, ok := existingStatuses[stepKey]; !ok || status != common.StepCompleted {
				segmentsToProcess = append(segmentsToProcess, i)
			}
		}

		// 3. Run the summary generation for the remaining segments in parallel
		resultsChan := make(chan result, len(segmentsToProcess))
		var wg sync.WaitGroup

		// Thread-safe map to capture actual errors from workers
		var processingErrors sync.Map

		// Goroutine to collect results and update metadata in batches
		var updateWg sync.WaitGroup
		updateWg.Add(1)
		var flushMutex sync.Mutex
		go func() {
			defer updateWg.Done()
			const batchSize = 5
			metadataUpdate := make(map[string]string)
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			flush := func() {
				flushMutex.Lock()
				defer flushMutex.Unlock()
				if len(metadataUpdate) > 0 {
					log.Printf("Persisting batch of %d segment summaries...", len(metadataUpdate))

					if _, err := config.UpdateGCSObjectMetadata(metadataUpdate); err != nil {
						log.Printf("Warning: failed to persist segment summary batch: %v", err)
					} else {
						metadataUpdate = make(map[string]string) //Only clear if write was successful
					}
				}
			}

			for {
				select {
				case res, ok := <-resultsChan:
					if !ok { // Channel closed
						flush()
						return
					}

					if res.err != nil {
						log.Printf("Error processing step %s: %v", res.stepKey, res.err)
						processingErrors.Store(res.stepKey, res.err)
						continue // Skip adding to metadata if failed
					}

					status := common.StepStatus{Output: res.output, Status: common.StepCompleted}
					if statusBytes, err := json.Marshal(status); err == nil {
						metadataUpdate[res.stepKey] = string(statusBytes)
					}
					if len(metadataUpdate) >= batchSize {
						flush()
					}
				case <-ticker.C:
					flush()
				}
			}
		}()

		const maxConcurrent = 10
		semaphore := make(chan struct{}, maxConcurrent)

		for _, segmentIndex := range segmentsToProcess {
			wg.Add(1)
			semaphore <- struct{}{} // Acquire a slot

			go func(segmentIndex int) {
				defer func() {
					<-semaphore // Release the slot
					wg.Done()
				}()
				stepKey := getSegmentSummaryStepKey(segmentIndex)
				log.Printf("Generating summary for segment %d", segmentIndex+1)
				output, err := get_segment_summary(config.GenaiRunConfig, contentSummaryObj, inputValues[common.CONTENT_TYPE_STEP], segmentIndex)
				resultsChan <- result{stepKey: stepKey, output: output, err: err}
			}(segmentIndex)
		}
		wg.Wait()
		close(resultsChan)
		updateWg.Wait() // Wait for the final metadata update to complete

		var anyErr error
		processingErrors.Range(func(key, value any) bool {
			anyErr = fmt.Errorf("step %v failed: %w", key, value.(error))
			return false // stop iteration
		})
		if anyErr != nil {
			return "", anyErr
		}

		segmentSummaryStepKeys := make([]string, len(contentSummaryObj.SegmentTimeStamps))
		for i := range contentSummaryObj.SegmentTimeStamps {
			segmentSummaryStepKeys[i] = getSegmentSummaryStepKey(i)
		}

		segmentSummaryStepsStatus := config.BasicRunConfig.GetStepsStatus(segmentSummaryStepKeys)

		for i := range contentSummaryObj.SegmentTimeStamps {
			stepKey := getSegmentSummaryStepKey(i)
			status, ok := segmentSummaryStepsStatus[stepKey]
			if !ok {
				return "", fmt.Errorf("segment summary step %s not found", stepKey)
			}
			if status != common.StepCompleted {
				return "", fmt.Errorf("segment summary step %s not completed", stepKey)
			}
		}

		return fmt.Sprintf("summary generated for %d segments", len(contentSummaryObj.SegmentTimeStamps)), nil
	}
}

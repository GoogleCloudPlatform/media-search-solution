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
	"os"
	"time"
)

const (
	// FileCheckRetries is the number of times to check for a file's existence.
	FileCheckRetries = 5
	// FileCheckDelay is the time to wait between file existence checks.
	FileCheckDelay = 10 * time.Second
)

// WaitForFile polls for the existence of a file with retries.
// This is useful when waiting for a file to appear in a GCS FUSE mounted directory.
func WaitForFile(filePath string) error {
	var err error
	for i := range FileCheckRetries {
		if _, err = os.Stat(filePath); err == nil {
			return nil
		}
		log.Printf("waiting for file to appear: %s, attempt %d/%d", filePath, i+1, FileCheckRetries)
		time.Sleep(FileCheckDelay)
	}
	return fmt.Errorf("file: %s not found after several retries. Error: %w", filePath, err)
}

// WaitForFileUpdate polls for the existence and recent modification of a file.
// This is useful for waiting for a file to be updated in a GCS FUSE mounted directory.
func WaitForFileUpdate(localFile string, recentThreshold time.Duration) {
	// it can take some time to sync the file from the bucket to the local filesystem.
	// We check for the file's existence and modification time to ensure we have the latest version.
	for i := range FileCheckRetries {
		fileInfo, err := os.Stat(localFile)
		if err == nil {
			if time.Since(fileInfo.ModTime()) < recentThreshold {
				log.Printf("Configuration file %s has been updated recently.", localFile)
				return
			}
		}
		log.Printf("waiting for configuration file to be updated: %s, attempt %d/%d", localFile, i+1, FileCheckRetries)
		time.Sleep(FileCheckDelay)
	}
	log.Printf("Configuration file %s not updated after several retries. Proceeding with existing config.", localFile)
}

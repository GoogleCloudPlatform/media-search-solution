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

package test

import (
	"log"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cloud"
)

type StateManager struct {
	config *cloud.Config
}

var state = &StateManager{}

func HandleErr(err error, t *testing.T) {
	if err != nil {
		t.Errorf("Error reading config file: %v", err)
	}
}

func GetTestHighResMessageText() string {
	return `{
  "kind": "storage#object",
  "id": "media_high_res_resources/test-trailer-001.mp4/1728615848664286",
  "selfLink": "https://www.googleapis.com/storage/v1/b/media_high_res_resources/o/test-trailer-001.mp4",
  "name": "test-trailer-001.mp4",
  "bucket": "media_high_res_resources",
  "generation": "1728615848664286",
  "metageneration": "1",
  "contentType": "video/mp4",
  "timeCreated": "2024-10-11T03:04:08.672Z",
  "updated": "2024-10-11T03:04:08.672Z",
  "storageClass": "STANDARD",
  "timeStorageClassUpdated": "2024-10-11T03:04:08.672Z",
  "size": "259348037",
  "md5Hash": "67c1rAU+1RYZzK5zp8iBkA==",
  "mediaLink": "https://storage.googleapis.com/download/storage/v1/b/media_high_res_resources/o/test-trailer-001.mp4?generation=1728615848664286&alt=media",
  "metadata": { "touch": "18" },
  "crc32c": "IYeSTw==",
  "etag": "CN658+yrhYkDEAE="
	}`
}

func GetTestLowResMessageText() string {
	return `{
  "kind": "storage#object",
  "id": "media_low_res_resources/test-trailer-001.mp4/1728615848664286",
  "selfLink": "https://www.googleapis.com/storage/v1/b/media_low_res_resources/o/test-trailer-001.mp4",
  "name": "test-trailer-001.mp4",
  "bucket": "media_low_res_resources",
  "generation": "1728615848664286",
  "metageneration": "1",
  "contentType": "video/mp4",
  "timeCreated": "2024-10-11T03:04:08.672Z",
  "updated": "2024-10-11T03:04:08.672Z",
  "storageClass": "STANDARD",
  "timeStorageClassUpdated": "2024-10-11T03:04:08.672Z",
  "size": "259348037",
  "md5Hash": "67c1rAU+1RYZzK5zp8iBkA==",
  "mediaLink": "https://storage.googleapis.com/download/storage/v1/b/media_low_res_resources/o/test-trailer-001.mp4?generation=1728615848664286&alt=media",
  "metadata": { "touch": "18" },
  "crc32c": "IYeSTw==",
  "etag": "CN658+yrhYkDEAE="
}
`
}

func SetupOS() (err error) {
	err = os.Setenv(cloud.EnvConfigFilePrefix, "configs")
	if err != nil {
		return err
	}
	err = os.Setenv(cloud.EnvConfigRuntime, "test")
	return err
}

func GetConfig() *cloud.Config {
	if state.config == nil {
		err := SetupOS()
		if err != nil {
			log.Fatalf("failed to setup environment for test: %v\n", err)
		}
		// Create a default cloud config
		config := cloud.NewConfig()
		// Load it from the TOML files
		cloud.LoadConfig(&config)
		state.config = config
	}
	return state.config
}

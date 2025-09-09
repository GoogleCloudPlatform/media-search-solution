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

package cloud

// GetGCSObjectName returns a placeholder string for a GCS object name.
func GetGCSObjectName() string {
	return "__GCS__OBJ__"
}

// GCSPubSubNotification is the structure of a message received from a
// Google Cloud Storage (GCS) Pub/Sub notification. It contains metadata
// about a change to an object in a GCS bucket.
type GCSPubSubNotification struct {
	Kind                    string                 `json:"kind"`
	ID                      string                 `json:"id"`
	SelfLink                string                 `json:"selfLink"`
	Name                    string                 `json:"name"`
	Bucket                  string                 `json:"bucket"`
	Generation              string                 `json:"generation"`
	MetaGeneration          string                 `json:"metageneration"`
	ContentType             string                 `json:"contentType"`
	TimeCreated             string                 `json:"timeCreated"`
	Updated                 string                 `json:"updated"`
	StorageClass            string                 `json:"storageClass"`
	TimeStorageClassUpdated string                 `json:"timeStorageClassUpdated"`
	Size                    string                 `json:"size"`
	MD5Hash                 string                 `json:"md5Hash"`
	MediaLink               string                 `json:"mediaLink"`
	MetaData                map[string]interface{} `json:"metadata"`
	Crc32c                  string                 `json:"crc32c"`
	ETag                    string                 `json:"etag"`
}

// GCSObject is a simplified representation of a Google Cloud Storage (GCS)
// object. It contains the bucket name, object name, and MIME type of the object.
type GCSObject struct {
	Bucket   string
	Name     string
	MIMEType string
}

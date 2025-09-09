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

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func FileUpload(r *gin.RouterGroup) {
	config := GetConfig()

	upload := r.Group("/uploads")
	{
		upload.POST("", func(c *gin.Context) {
			form, err := c.MultipartForm()
			if err != nil {
				c.Status(400)
				return
			}
			files := form.File["files"]
			bucket := state.cloud.StorageClient.Bucket(config.Storage.HiResInputBucket)

			for _, file := range files {
				localPath := filepath.Join(os.TempDir(), file.Filename)
				err := c.SaveUploadedFile(file, localPath)
				if err != nil {
					log.Println(err)
					c.Status(400)
					return
				}
				content, err := os.ReadFile(localPath)
				if err != nil {
					log.Println(err)
					c.Status(400)
					return
				}
				wc := bucket.Object(file.Filename).NewWriter(c)
				wc.ContentType = "video/mp4"
				_, err = wc.Write(content)
				if err != nil {
					c.Status(500)
					log.Printf("failed to write file to bucket: %v\n", err)
					return
				}
				err = wc.Close()
				if err != nil {
					log.Printf("failed to close bucket handle: %v\n", err)
				}
				err = os.Remove(localPath)
				if err != nil {
					log.Printf("failed to remove file from server: %v\n", err)
				}
			}
			c.Status(200)
		})
	}
}

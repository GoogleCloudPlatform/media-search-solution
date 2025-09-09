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
	"strconv"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/model"
	"github.com/gin-gonic/gin"
)

func MediaRouter(r *gin.RouterGroup) {
	media := r.Group("/media")
	{
		media.GET("", func(c *gin.Context) {
			query := c.Query("s")
			count, err := strconv.Atoi(c.DefaultQuery("count", "5"))
			if err != nil {
				count = 5
			}
			if len(query) == 0 {
				c.Status(404)
				return
			}
			sceneResults, err := state.searchService.FindScenes(c, query, count)

			if err != nil {
				c.Status(404)
				log.Println(err)
				return
			}

			out := make(map[string]*model.Media, 0)

			// Convert the results into a map driven by the media id
			for _, r := range sceneResults {
				var med *model.Media
				if m, ok := out[r.MediaId]; !ok {
					m, err := state.mediaService.Get(c, r.MediaId)
					if err != nil {
						log.Print(err)
						c.Status(400)
						return
					}
					// Clear the scenes
					m.Scenes = make([]*model.Scene, 0)
					out[r.MediaId] = m
					med = m
				} else {
					med = m
				}

				s, err := state.mediaService.GetScene(c, r.MediaId, r.SequenceNumber)
				if err != nil {
					c.Status(400)
					return
				}
				med.Scenes = append(med.Scenes, s)
			}
			// Reduce
			results := make([]*model.Media, 0)
			for _, v := range out {
				results = append(results, v)
			}
			c.JSON(200, results)
		})

		media.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			out, err := state.mediaService.Get(c, id)
			if err != nil {
				c.Status(404)
				return
			}
			c.JSON(200, out)
		})

		media.GET("/:id/scenes/:scene_id", func(c *gin.Context) {
			id := c.Param("id")
			sceneId, err := strconv.Atoi(c.Param("scene_id"))
			if err != nil {
				c.Status(400)
				return
			}
			out, err := state.mediaService.GetScene(c, id, sceneId)
			if err != nil {
				c.Status(404)
				return
			}
			c.JSON(200, out)
		})
	}
}

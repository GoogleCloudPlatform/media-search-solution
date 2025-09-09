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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GoogleCloudPlatform/media-search-solution/pkg/telemetry"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	telemetry.SetupLogging()

	log.Print("Logging initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := telemetry.SetupOpenTelemetry(ctx, GetConfig())
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Tracing initialized")

	InitState(ctx)
	log.Println("Initialized State")

	r := gin.Default()

	r.Use(otelgin.Middleware("media-search-server"))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	// Create the "/api/v1" group
	apiV1 := r.Group("/api/v1")
	{
		// Register "/api/v1/media" end-points
		MediaRouter(apiV1)
		// Register "/api/v1/uploads"
		FileUpload(apiV1)
	}

	// serving the front-end asset
	staticPath := "web/apps/media-search/dist"
	r.Static("/assets", staticPath+"/assets")
	r.StaticFile("/favicon.ico", staticPath+"/favicon.ico")
	r.NoRoute(func(c *gin.Context) {
		c.File(staticPath + "/index.html")
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Print("Server Ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	// Shutdown in 60 seconds if shutdown
	oCtx, oCancel := context.WithTimeout(ctx, 60*time.Second)
	defer oCancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	select {
	case <-oCtx.Done():
		log.Println("Timeout, failed to shutdown gracefully")
	}
	log.Println("Server exiting")
}

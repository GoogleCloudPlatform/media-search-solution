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
// Author:  kingman (Charlie Wang)

terraform {
  required_version = ">= 0.12"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.5.0"
    }
  }
}

data "google_storage_project_service_account" "gcs_account" {
}

resource "google_pubsub_topic" "media_config_update_events" {
  name = "media_config_update_events"
  message_storage_policy {
    allowed_persistence_regions = [var.region]
  }
}

resource "google_storage_bucket" "media_config_resources" {
  name                        = var.config_bucket
  location                    = var.region
  uniform_bucket_level_access = true
  force_destroy               = true
  public_access_prevention    = "enforced"
  versioning {
    enabled = true
  }
}

resource "google_pubsub_subscription" "media_config_update_events_subscription" {
  name                    = "media_config_update_events_subscription"
  topic                   = google_pubsub_topic.media_config_update_events.id
  enable_message_ordering = true
  ack_deadline_seconds    = 300
}

resource "google_storage_notification" "media_config_resource_notifications" {
  bucket         = google_storage_bucket.media_config_resources.name
  payload_format = "JSON_API_V1"
  topic          = google_pubsub_topic.media_config_update_events.id
  event_types    = ["OBJECT_FINALIZE"]
  depends_on = [google_pubsub_topic_iam_binding.pubsub_invoker]
}

resource "google_pubsub_topic_iam_binding" "pubsub_invoker" {
  topic  = google_pubsub_topic.media_config_update_events.id
  role   = "roles/pubsub.publisher"
  members = ["serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"]
}
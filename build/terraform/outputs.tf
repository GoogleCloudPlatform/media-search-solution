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

output "project_id" {
  description = "The ID of the Google Cloud project where resources are created."
  value       = var.project_id
}

output "high_res_bucket" {
  description = "The name of the high resolution media bucket."
  value       = var.high_res_bucket
}

output "low_res_bucket" {
  description = "The name of the low resolution media bucket."
  value       = var.low_res_bucket
}

output "config_bucket" {
  description = "The name of the configuration files bucket."
  value       = var.config_bucket
}

output "cloud_run_service_name" {
  description = "The name of the Cloud Run service to deploy."
  value       = local.solution_prefix
}

output "cloud_run_region" {
  description = "The region where the Cloud Run service should be deployed."
  value       = var.region
}

output "service_account_email" {
  description = "The email of the service account for the Cloud Run service."
  value       = module.media_search_service_account.email
}

output "container_image_uri" {
  description = "The URI of the container image to deploy."
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker-repo.name}/${local.solution_prefix}"
}
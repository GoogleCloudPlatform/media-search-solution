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
// Author:  rrmcguinness (Ryan McGuinness)
//          kingman (Charlie Wang)

terraform {
  required_version = ">= 0.12"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.5.0"
    }
  }

  provider_meta "google" {
    module_name = "cloud-solutions/media-search-solution-deploy-v1.0.0"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

locals {
  solution_prefix = "media-search"
}

resource "google_artifact_registry_repository" "docker-repo" {
  format        = "DOCKER"
  location      = var.region
  repository_id = "${local.solution_prefix}-repo"
  description   = "Docker containers"
}

module "low_res_resources" {
  source         = "./modules/low_res"
  region         = var.region
  low_res_bucket = var.low_res_bucket
}

module "high_res_resources" {
  source          = "./modules/high_res"
  region          = var.region
  high_res_bucket = var.high_res_bucket
}

module "bigquery" {
  source = "./modules/bigquery"
  region = var.region
}

resource "local_file" "deployment_configuration" {
  filename = "${path.root}/../../configs/.env.local.toml"

  content = templatefile("${path.module}/../../configs/example_.env.local.toml", {
    project_id            = var.project_id
    high_res_input_bucket = var.high_res_bucket
    low_res_output_bucket = var.low_res_bucket
  })
}

module "config_resources" {
  source        = "./modules/config_res"
  region        = var.region
  config_bucket = var.config_bucket
}

module "cloud_build_account" {
  source     = "github.com/terraform-google-modules/terraform-google-service-accounts?ref=a11d4127eab9b51ec9c9afdaf51b902cd2c240d9" #commit hash of version 4.0.0
  names      = ["cloud-build"]
  project_id = var.project_id
  project_roles = [
    "${var.project_id}=>roles/logging.logWriter",
    "${var.project_id}=>roles/storage.admin",
    "${var.project_id}=>roles/artifactregistry.writer",
    "${var.project_id}=>roles/run.developer",
  ]
  display_name = "Cloud Build Service Account"
  description  = "specific custom service account for Cloud Build"
}

module "gcloud_build_app" {
  source = "github.com/terraform-google-modules/terraform-google-gcloud?ref=db25ab9c0e9f2034e45b0034f8edb473dde3e4ff" # commit hash of version 3.5.0

  create_cmd_entrypoint = "gcloud"
  create_cmd_body       = <<-EOT
    builds submit "${path.module}/../.." \
      --project ${var.project_id} \
      --region ${var.region} \
      --default-buckets-behavior regional-user-owned-bucket \
      --tag "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker-repo.name}/${local.solution_prefix}" \
      --service-account "projects/${var.project_id}/serviceAccounts/${module.cloud_build_account.email}"
  EOT
  enabled               = true
}

module "media_search_service_account" {
  source     = "github.com/terraform-google-modules/terraform-google-service-accounts?ref=a11d4127eab9b51ec9c9afdaf51b902cd2c240d9" #commit hash of version 4.0.0
  project_id = var.project_id
  names      = ["media-search-sa"]
  project_roles = [
    "${var.project_id}=>roles/storage.objectAdmin",
    "${var.project_id}=>roles/bigquery.user",
    "${var.project_id}=>roles/bigquery.dataEditor",
    "${var.project_id}=>roles/aiplatform.user",
    "${var.project_id}=>roles/pubsub.subscriber",
    "${var.project_id}=>roles/cloudtrace.agent",
    "${var.project_id}=>roles/monitoring.metricWriter",
  ]
  display_name = "Media Search Web Service Account"
  description  = "specific custom service account for Web APP"
}

#!/usr/bin/env bash
#
# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Author: kingman (Charlie Wang)

# This script cleans up all resources associated with a specific media file.
# It deletes the file from both high-resolution and low-resolution
# Cloud Storage buckets and removes corresponding entries from BigQuery tables.

set -euo pipefail

# --- Helper Functions ---
info() {
  echo "[INFO]    $*"
}

error() {
  echo "[ERROR]   $*" >&2
  exit 1
}

# --- Main Logic ---
main() {
  if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <media_file_name>"
    echo "  <media_file_name>: The name of the video file to clean up (e.g., 'my_video.mp4')."
    exit 1
  fi

  local media_file_name="$1"
  info "Starting cleanup for media file: ${media_file_name}"

  local terraform_dir="build/terraform"
  if [[ ! -d "${terraform_dir}" ]]; then
      error "Terraform directory not found at '${terraform_dir}'. Please run this script from the project root directory."
  fi

  # Get configuration from Terraform outputs
  info "Fetching configuration from Terraform..."
  local project_id
  project_id=$(terraform -chdir="${terraform_dir}" output -raw project_id)
  if [[ -z "${project_id}" ]]; then
      error "Could not retrieve project_id from Terraform outputs."
  fi

  local high_res_bucket
  high_res_bucket=$(terraform -chdir="${terraform_dir}" output -raw high_res_bucket)
  if [[ -z "${high_res_bucket}" ]]; then
      error "Could not retrieve high_res_bucket from Terraform outputs."
  fi

  local low_res_bucket
  low_res_bucket=$(terraform -chdir="${terraform_dir}" output -raw low_res_bucket)
  if [[ -z "${low_res_bucket}" ]]; then
      error "Could not retrieve low_res_bucket from Terraform outputs."
  fi

  local bq_dataset
  bq_dataset="media_ds"
  info "Using BigQuery dataset: ${bq_dataset}"

    # --- BigQuery Cleanup ---
  info "Cleaning up BigQuery records..."

  local table
  table="segment_embeddings"
  info "Deleting records from table: ${table}"
  bq query --project_id="${project_id}" --use_legacy_sql=false \
    "DELETE FROM \`${project_id}.${bq_dataset}.${table}\` WHERE media_id IN (SELECT id FROM \`${project_id}.${bq_dataset}.media\` WHERE media_url LIKE '%${media_file_name}')"
  table="media"
  bq query --project_id="${project_id}" --use_legacy_sql=false \
    "DELETE FROM \`${project_id}.${bq_dataset}.${table}\` WHERE media_url LIKE '%${media_file_name}'"

  # --- Cloud Storage Cleanup ---
  info "Cleaning up Cloud Storage files..."

  # High-resolution file
  local high_res_uri="gs://${high_res_bucket}/${media_file_name}"
  if gsutil -q stat "${high_res_uri}"; then
    info "Deleting high-resolution file: ${high_res_uri}"
    gsutil rm "${high_res_uri}"
  else
    info "High-resolution file not found, skipping: ${high_res_uri}"
  fi

  # Low-resolution file
  local low_res_uri="gs://${low_res_bucket}/${media_file_name}"
  if gsutil -q stat "${low_res_uri}"; then
    info "Deleting low-resolution file: ${low_res_uri}"
    gsutil rm "${low_res_uri}"
  else
    info "Low-resolution file not found, skipping: ${low_res_uri}"
  fi

  info "Cleanup for '${media_file_name}' completed successfully."
}

main "$@"

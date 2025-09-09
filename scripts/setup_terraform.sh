#!/usr/bin/env bash

# Copyright 2024 Google, LLC
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
#         gfilicetti (Gino Filicetti)

# This script automates the initial setup required before running Terraform.
# It ensures that Terraform and gcloud are installed, logs into Google Cloud,
# sets the correct project, and enables all necessary APIs and permissions.

set -e

# Determine the project root directory, assuming the script is in a subdirectory of the root
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")

echo "--- Media Search Solution Pre-Terraform Setup ---"
echo

# --- Tool Checks ---
echo "🔎 Checking for required tools..."

if ! command -v terraform &> /dev/null; then
  echo "❌ ERROR: Terraform could not be found. Please install Terraform."
  echo "See: https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/install-cli"
  exit 1
fi


if ! command -v gcloud &> /dev/null; then
  echo "❌ ERROR: gcloud could not be found. Please install the Google Cloud SDK."
  echo "See: https://cloud.google.com/sdk/docs/install"
  exit 1
fi



read -p "Enter your Google Cloud project ID: " -r PROJECT_ID
if [ -z "$PROJECT_ID" ]; then
  echo "❌ ERROR: Project ID cannot be empty."
  exit 1
fi

scripts/gcloud_login.sh

echo
echo "Setting active gcloud project to '$PROJECT_ID'..."
gcloud config unset billing/quota_project
gcloud config set project "${PROJECT_ID}"
gcloud auth application-default set-quota-project "$PROJECT_ID"
echo

# --- API Enablement ---
APIS_TO_ENABLE=(
  "aiplatform.googleapis.com"
  "artifactregistry.googleapis.com"
  "cloudbuild.googleapis.com"
  "cloudresourcemanager.googleapis.com"
  "iam.googleapis.com"
  "iap.googleapis.com"
  "pubsub.googleapis.com"
  "run.googleapis.com"
  "storage.googleapis.com"
)

echo "Enabling necessary Google Cloud APIs. This may take a few minutes..."
for i in "${APIS_TO_ENABLE[@]}"; do
  gcloud services enable "$i" --project="$PROJECT_ID"
done
echo


echo "Creating Vertex AI service agent identity..."
gcloud beta services identity create --service=aiplatform.googleapis.com --project="$PROJECT_ID"

echo "Waiting 60 seconds for the service agent to be created and propagated..."
sleep 60

echo "Retrieving project number..."
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
if [ -z "$PROJECT_NUMBER" ]; then
    echo "❌ ERROR: Could not retrieve project number for project '$PROJECT_ID'."
    exit 1
fi

SERVICE_AGENT="service-${PROJECT_NUMBER}@gcp-sa-aiplatform.iam.gserviceaccount.com"
echo "Vertex AI Service Agent: $SERVICE_AGENT"

SERVICE_AGENT_ROLES=(
  "roles/aiplatform.serviceAgent"
  "roles/storage.objectAdmin"
  "roles/bigquery.dataViewer"
  "roles/bigquery.jobUser"
  "roles/pubsub.admin"
)

echo "Assigning IAM roles to the Vertex AI service agent..."
for role in "${SERVICE_AGENT_ROLES[@]}"; do

  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_AGENT" \
    --role="$role" \
    --quiet
done
echo "Service Agent IAM roles assigned."
echo

# --- Terraform Variables File ---
TF_VARS_FILE="$PROJECT_ROOT/build/terraform/terraform.tfvars"
TF_VARS_EXAMPLE_FILE="$PROJECT_ROOT/build/terraform/terraform.tfvars.example"

echo "Checking for Terraform variables file..."
if [ ! -f "$TF_VARS_FILE" ]; then
  echo "Terraform variables file not found. Creating from example..."
  cp "$TF_VARS_EXAMPLE_FILE" "$TF_VARS_FILE"
  echo "✅ Created '$TF_VARS_FILE'. Please edit this file with your project settings."
else
  echo "✅ Terraform variables file '$TF_VARS_FILE' already exists."
fi
echo

# --- Final Message ---
echo "🎉 Pre-Terraform setup is complete!"
echo "You can now edit '$TF_VARS_FILE' and then run 'terraform init' and 'terraform apply' from the 'build/terraform' directory."

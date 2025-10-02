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

# This script automates the deployment of the Media Search service to Cloud Run.
# It assumes that you have already run `terraform apply` and have the necessary
# Terraform outputs available.

set -e # Exit immediately if a command exits with a non-zero status.

# Determine the project root directory, assuming the script is in a subdirectory of the root
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
TERRAFORM_DIR="$PROJECT_ROOT/build/terraform"

# Retrieve Terraform outputs
PROJECT_ID=$(terraform -chdir="$TERRAFORM_DIR" output -raw project_id)
CLOUD_RUN_SERVICE_NAME=$(terraform -chdir="$TERRAFORM_DIR" output -raw cloud_run_service_name)
CLOUD_RUN_REGION=$(terraform -chdir="$TERRAFORM_DIR" output -raw cloud_run_region)
CONTAINER_IMAGE_URI=$(terraform -chdir="$TERRAFORM_DIR" output -raw container_image_uri)
SERVICE_ACCOUNT_EMAIL=$(terraform -chdir="$TERRAFORM_DIR" output -raw service_account_email)
HIGH_RES_BUCKET=$(terraform -chdir="$TERRAFORM_DIR" output -raw high_res_bucket)
LOW_RES_BUCKET=$(terraform -chdir="$TERRAFORM_DIR" output -raw low_res_bucket)
CONFIG_BUCKET=$(terraform -chdir="$TERRAFORM_DIR" output -raw config_bucket)

# Check if the variables are empty
if [ -z "$PROJECT_ID" ] || [ -z "$CLOUD_RUN_SERVICE_NAME" ] || [ -z "$CLOUD_RUN_REGION" ] || [ -z "$CONTAINER_IMAGE_URI" ] || [ -z "$SERVICE_ACCOUNT_EMAIL" ] || [ -z "$HIGH_RES_BUCKET" ] || [ -z "$LOW_RES_BUCKET" ]; then
  echo "ERROR: One or more Terraform output variables are not set. Ensure terraform apply was successful."
  exit 1
fi

PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')

# Undeploy the current Cloud Run service (if it exists)
echo "Checking for existing Cloud Run service..."
if gcloud run services describe "$CLOUD_RUN_SERVICE_NAME" --project "$PROJECT_ID" --region "$CLOUD_RUN_REGION" &>/dev/null; then
  echo "Undeploying existing Cloud Run service..."
  gcloud run services delete "$CLOUD_RUN_SERVICE_NAME" \
    --project "$PROJECT_ID" \
    --region "$CLOUD_RUN_REGION" \
    --quiet
else
    echo "No existing service to undeploy."
fi

if [ "$(gsutil -q stat gs://${CONFIG_BUCKET}/.env.toml ; echo $?)" = 0 ]; then
  echo ".env.toml file is already uploaded to gs://${CONFIG_BUCKET}"
else
  echo "Uploading .env.toml file to gs://${CONFIG_BUCKET}"
  gsutil cp "${PROJECT_ROOT}/configs/.env.toml" "gs://${CONFIG_BUCKET}/.env.toml"
fi

if [ "$(gsutil -q stat gs://${CONFIG_BUCKET}/.env.local.toml ; echo $?)" = 0 ]; then
  echo ".env.local.toml file is already uploaded to gs://${CONFIG_BUCKET}"
else
  echo "Uploading .env.local.toml file to gs://${CONFIG_BUCKET}"
  gsutil cp "${PROJECT_ROOT}/configs/.env.local.toml" "gs://${CONFIG_BUCKET}/.env.local.toml"
fi

# Deploy the service to Cloud Run
echo "Deploying service to Cloud Run..."
gcloud run deploy "$CLOUD_RUN_SERVICE_NAME" \
  --project "$PROJECT_ID" \
  --region "$CLOUD_RUN_REGION" \
  --image "$CONTAINER_IMAGE_URI" \
  --service-account "$SERVICE_ACCOUNT_EMAIL" \
  --add-volume name=high-res-bucket,type=cloud-storage,bucket="$HIGH_RES_BUCKET" \
  --add-volume-mount volume=high-res-bucket,mount-path=/mnt/"$HIGH_RES_BUCKET" \
  --add-volume name=low-res-bucket,type=cloud-storage,bucket="$LOW_RES_BUCKET" \
  --add-volume-mount volume=low-res-bucket,mount-path=/mnt/"$LOW_RES_BUCKET" \
  --add-volume name=config-bucket,type=cloud-storage,bucket="$CONFIG_BUCKET" \
  --add-volume-mount volume=config-bucket,mount-path=/mnt/"$CONFIG_BUCKET" \
  --set-env-vars GCP_CONFIG_PREFIX=/mnt/"$CONFIG_BUCKET" \
  --cpu=8 \
  --memory=8Gi \
  --no-cpu-throttling \
  --no-allow-unauthenticated \
  --timeout=3600 \
  --min 1

  # Create the IAP service agent
gcloud beta services identity create \
  --service=iap.googleapis.com \
  --project="$PROJECT_ID"

# Grant the cloud run invoker permission to the IAP service account:
gcloud run services add-iam-policy-binding "$CLOUD_RUN_SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$CLOUD_RUN_REGION" \
  --member="serviceAccount:service-${PROJECT_NUMBER}@gcp-sa-iap.iam.gserviceaccount.com" \
  --role='roles/run.invoker'

# Enable IAP on the Cloud Run service:
gcloud beta run services update "$CLOUD_RUN_SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$CLOUD_RUN_REGION" \
  --iap

echo "IAP enabled. Now you can grant access to users or domains."

# Grant access to the application via IAP
while true; do
  # Make sure we are reading from the terminal, even if stdin is redirected
  read -p "Add IAP policy for a 'user' or a 'domain'? (Enter 'user', 'domain', or 'quit' to finish): " -r policy_type < /dev/tty

  case "$policy_type" in
    user|domain)
      read -p "Enter the ${policy_type}'s identifier (e.g., user@example.com or example.com): " -r identifier < /dev/tty
      if [ -z "$identifier" ]; then
        echo "Identifier cannot be empty. Please try again."
        continue
      fi
      echo "Adding IAP policy for ${policy_type}: $identifier"
      gcloud beta iap web add-iam-policy-binding \
        --service="$CLOUD_RUN_SERVICE_NAME" \
        --resource-type=cloud-run \
        --project="$PROJECT_ID" \
        --region="$CLOUD_RUN_REGION" \
        --member="${policy_type}:${identifier}" \
        --role='roles/iap.httpsResourceAccessor'
        break
      ;;
    quit)
      echo "Finished adding IAP policies."
      break
      ;;
    *)
      echo "Invalid input. Please enter 'user', 'domain', or 'quit'."
      ;;
  esac
done

echo "IAP policies set. Now you can grant bucket permissions to users or domains."

# Grant access to the Cloud Storage buckets
while true; do
  # Make sure we are reading from the terminal, even if stdin is redirected
  read -p "Add Storage permissions for a 'user' or a 'domain'? (Enter 'user', 'domain', or 'quit' to finish): " -r policy_type < /dev/tty

  case "$policy_type" in
    user|domain)
      read -p "Enter the ${policy_type}'s identifier (e.g., user@example.com or example.com): " -r identifier < /dev/tty
      if [ -z "$identifier" ]; then
        echo "Identifier cannot be empty. Please try again."
        continue
      fi

      echo "Granting '${policy_type}:${identifier}' permission to upload to high-res bucket: gs://${HIGH_RES_BUCKET}"
      gcloud storage buckets add-iam-policy-binding "gs://${HIGH_RES_BUCKET}" \
        --project="$PROJECT_ID" \
        --member="${policy_type}:${identifier}" \
        --role="roles/storage.objectCreator"

      echo "Granting '${policy_type}:${identifier}' permission to view low-res bucket: gs://${LOW_RES_BUCKET}"
      gcloud storage buckets add-iam-policy-binding "gs://${LOW_RES_BUCKET}" \
        --project="$PROJECT_ID" \
        --member="${policy_type}:${identifier}" \
        --role="roles/storage.objectViewer"
      break
      ;;
    quit)
      echo "Finished adding Storage permissions."
      break
      ;;
    *)
      echo "Invalid input. Please enter 'user', 'domain', or 'quit'."
      ;;
  esac
done

echo "Cloud Run deployment and all configurations are complete!"
echo "you can now access your service at: https://$CLOUD_RUN_SERVICE_NAME-$PROJECT_NUMBER.$CLOUD_RUN_REGION.run.app"
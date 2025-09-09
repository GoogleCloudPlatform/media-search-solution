#!/usr/bin/env bash
#
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

# This script checks for and establishes gcloud user and application-default
# credentials.

set -e

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "gcloud command could not be found. Please install the Google Cloud SDK." >&2
    exit 1
fi

# Check if the user is logged in to gcloud CLI.
echo "Checking gcloud user authentication..."
if gcloud projects list --quiet &>/dev/null; then
    echo "gcloud user is already authenticated."
else
    echo "gcloud user not authenticated. Please follow the prompts to log in."
    gcloud auth login
fi

echo

# Check if Application Default Credentials (ADC) are set up.
echo "Checking Application Default Credentials (ADC)..."
if gcloud auth application-default print-access-token --quiet &>/dev/null; then
    echo "Application Default Credentials are set."
else
    echo "Application Default Credentials not set. Please follow the prompts to log in."
    gcloud auth application-default login
fi

echo
echo "gcloud authentication checks are complete."

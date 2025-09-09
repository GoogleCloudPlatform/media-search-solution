# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Author:  rrmcguinness (Ryan McGuinness)

terraform {
  required_version = ">= 0.12"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.5.0"
    }
  }
}

resource "google_compute_instance_template" "tpl" {
  name_prefix  = "media-pipeline-nodejs-instance-template-"
  machine_type = "e2-medium"
  zone         = var.region

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-2204-lts"
    disk_size_gb = 500
  }

  network_interface {
    network = "default"
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  metadata = {
    block-project-ssh-keys = true
  }

  metadata_startup_script = <<EOF
#!/bin/bash

# Update package lists
sudo apt-get update

# Install FFmpeg
sudo apt-get install -y ffmpeg

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Copy your Node.js program (replace with your actual implementation)
# Example:
# sudo gsutil cp gs://your-bucket/your-program.js /home/your-user/

# Start your Node.js program (replace with your actual implementation)
# Example:
# node /home/your-user/your-program.js
EOF
}

resource "google_compute_region_instance_group_manager" "mig" {
  name = "media-pipeline-nodejs-mig"
  version {
    instance_template = google_compute_instance_template.tpl.id
    name              = "primary"
  }
  base_instance_name  = "media-pipeline-nodejs-instance"
  zone                = var.region
  target_size         = 1
}

resource "google_compute_region_autoscaler" "autoscaler" {
  name    = "mdeia-pipeline-nodejs-autoscaler"
  target  = google_compute_region_instance_group_manager.mig.id
  zone = var.region

  autoscaling_policy {
    cpu_utilization {
      predictive_method = "NONE"
      target            = 0.6
    }
    min_replicas = 1
    max_replicas = 10
    mode         = "ON"
  }
}
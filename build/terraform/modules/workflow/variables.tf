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
// Author: kingman (Charlie Wang)

variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "proxy_generator_container" {
  type = string
}

variable "media_analysis_container" {
  type = string
}

variable "low_res_bucket" {
  type = string
}

variable "config_bucket" {
  type = string
}

variable "high_res_bucket" {
  type = string
}


variable "media_search_service_account_email" {
  type = string
}
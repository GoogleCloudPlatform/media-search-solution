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
# Author: kingman (Charlie Wang)

# Stage 1: The Builder
# This stage uses a Go image, installs Bazel, and builds our application.
FROM golang:1.21-bookworm AS builder


# Install Bazel and other dependencies

RUN apt-get update && apt-get install -y --no-install-recommends apt-transport-https curl gnupg build-essential && \
    curl -fsSL https://bazel.build/bazel-release.pub.gpg | gpg --dearmor > /etc/apt/trusted.gpg.d/bazel.gpg && \
    echo "deb [arch=amd64] https://storage.googleapis.com/bazel-apt stable jdk1.8" | tee /etc/apt/sources.list.d/bazel.list && \
    apt-get update && \
    apt-get install -y bazel && \
    # Clean up apt lists to reduce image size
    rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /src

COPY . .

# Build the 'generate_proxy' go_binary target using Bazel.
RUN bazel build //analyze/steps/proxy:generate_proxy --config=linux

# Stage 2: The Final Image
# This stage uses a minimal "distroless" image for a small and secure footprint.
FROM gcr.io/distroless/base-debian12

# Set the working directory
WORKDIR /app

# Copy the ffmpeg binary and its libraries.
COPY --from=builder /src/bazel-bin/bin/ bin/

# Copy the compiled 'generate_proxy' binary from the builder stage.
COPY --from=builder /src/bazel-bin/analyze/steps/proxy/generate_proxy_/generate_proxy .

ENTRYPOINT ["/app/generate_proxy"]

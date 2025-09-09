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
# Using 'bookworm' which is Debian 12
RUN apt-get update && apt-get install -y --no-install-recommends apt-transport-https curl gnupg build-essential && \
    curl -fsSL https://bazel.build/bazel-release.pub.gpg | gpg --dearmor > /etc/apt/trusted.gpg.d/bazel.gpg && \
    echo "deb [arch=amd64] https://storage.googleapis.com/bazel-apt stable jdk1.8" | tee /etc/apt/sources.list.d/bazel.list && \
    apt-get update && \
    apt-get install -y bazel && \
    # Clean up apt lists to reduce image size
    rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /src

# Copy the entire project context into the container.
# For more optimized builds, you could copy only necessary files (WORKSPACE, BUILD, go.mod, etc.) first.
COPY . .

# Build the api_server binary for a Linux environment.
# This command produces a statically-linked binary.
RUN bazel build //web/apps/api_server --config=linux


# Stage 2: The Final Image
# This stage uses a minimal "distroless" image for a small and secure footprint.
FROM gcr.io/distroless/base-debian12

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage.
# The path comes from how Bazel structures its output: bazel-bin/path/to/package/target_
COPY --from=builder /src/bazel-bin/web/apps/api_server/api_server_/api_server .

# Copy the static assets built by the media-search app.
COPY --from=builder /src/bazel-bin/web/apps/media-search/dist/ web/apps/media-search/dist/

# Copy the ffmpeg binary and its libraries.
COPY --from=builder /src/bazel-bin/bin/ bin/

# The api_server listens on port 8080
EXPOSE 8080

# The command to run when the container starts.
ENTRYPOINT ["/app/api_server"]

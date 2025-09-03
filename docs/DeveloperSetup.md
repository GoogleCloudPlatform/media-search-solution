# Developer Guide

## Developer Tools

Use the following instructions to set up a development environment:
* [Workstation Setup](WorkstationSetup.md)
* [Setting Up IntelliJ](SettingUpIntelliJ.md)
* [Setting Up Visual Studio Code](SettingUpVisualStudioCode.md)

## Check configuration file
The `configs/.env.local.toml` file is automatically generated after successfully running `terraform apply` in the `build/terraform` directory (see the [Deployment Guide](../README.md#create-infrastructure-resources-on-google-cloud) for details).

To verify the configuration, check that the following variables are properly set within the `[application]` and `[storage]` sections:

* `[application]` - `google_project_id`
* `[storage]` - `high_res_input_bucket`
* `[storage]` - `low_res_output_bucket`

```sh
cat configs/.env.local.toml
```

## Set up GCS Fuse
1. Follow the official [Cloud Storage FUSE installation guide](https://cloud.google.com/storage/docs/cloud-storage-fuse/install) to install it on your machine. Ensure you have also authenticated correctly (e.g., via `gcloud auth application-default login`).

1. Run the following script mounts the high-resolution and low-resolution buckets to a local directory (`~/media-search-mnt`).
**Note**: This script should be run from the project's root directory and requires that you have successfully run `terraform apply` in the `build/terraform` directory.

    ```sh
    HIGH_RES_BUCKET=$(terraform -chdir=build/terraform output -raw high_res_bucket)
    LOW_RES_BUCKET=$(terraform -chdir=build/terraform output -raw low_res_bucket)
    CONFIG_BUCKET=$(terraform -chdir=build/terraform output -raw config_bucket)
    AUDIO_BUCKET=$(terraform -chdir=build/terraform output -raw audio_bucket)
    ROOT_MOUNT_DIR="$HOME/media-search-mnt"
    HIGH_RES_MOUNT_POINT="$ROOT_MOUNT_DIR/$HIGH_RES_BUCKET"
    LOW_RES_MOUNT_POINT="$ROOT_MOUNT_DIR/$LOW_RES_BUCKET"
    CONFIG_MOUNT_POINT="$ROOT_MOUNT_DIR/$CONFIG_BUCKET"
    AUDIO_MOUNT_POINT="$ROOT_MOUNT_DIR/$AUDIO_BUCKET"
    mkdir -p "$HIGH_RES_MOUNT_POINT"
    mkdir -p "$LOW_RES_MOUNT_POINT"
    mkdir -p "$CONFIG_MOUNT_POINT"
    mkdir -p "$AUDIO_MOUNT_POINT"
    fusermount -u "$HIGH_RES_MOUNT_POINT"
    fusermount -u "$LOW_RES_MOUNT_POINT"
    fusermount -u "$CONFIG_MOUNT_POINT"
    fusermount -u "$AUDIO_MOUNT_POINT"
    gcsfuse "$HIGH_RES_BUCKET" "$HIGH_RES_MOUNT_POINT"
    gcsfuse "$LOW_RES_BUCKET" "$LOW_RES_MOUNT_POINT"
    gcsfuse "$CONFIG_BUCKET" "$CONFIG_MOUNT_POINT"
    gcsfuse "$AUDIO_BUCKET" "$AUDIO_MOUNT_POINT"
    ```

1. Next, you need to inform the application where to find the GCS Fuse mount point. This is done by adding the `gcs_fuse_mount_point` setting to your local configuration file (`configs/.env.local.toml`). The following command automates this update. It adds the configuration under the `[storage]` section
```sh
MOUNT_POINT_PATH=$(cd "$HOME/media-search-mnt" && pwd)
sed -i "/low_res_output_bucket/a gcs_fuse_mount_point = \"$MOUNT_POINT_PATH\"" configs/.env.local.toml
```
## Running the Demo Locally

```shell
# The following command combines two commands to simplify how the demo can be run
# bazel run //web/apps/api_server and bazel run //web/apps/media-search:start  
bazel run //:demo --action_env=NODE_ENV=development
```

## Building

```shell

# Build all targets
bazel build //...

# Build a specific target (The pipeline target in the pkg directory)
bazel build //pkg/model

# Build all targets in a specific package
bazel build //pkg/...

# Testing
bazel test //...

# Running Commands

bazel run //cmd:pipeline

# Cleaning
bazel clean

# Clean all cache
bazel clean --expunge

# Update Build Files and Dependencies
# Used when getting "missing strict dependency errors"
bazel run //:gazelle
```
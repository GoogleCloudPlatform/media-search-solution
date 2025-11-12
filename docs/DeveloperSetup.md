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
    ROOT_MOUNT_DIR="$HOME/media-search-mnt"
    HIGH_RES_MOUNT_POINT="$ROOT_MOUNT_DIR/$HIGH_RES_BUCKET"
    LOW_RES_MOUNT_POINT="$ROOT_MOUNT_DIR/$LOW_RES_BUCKET"
    CONFIG_MOUNT_POINT="$ROOT_MOUNT_DIR/$CONFIG_BUCKET"
    mkdir -p "$HIGH_RES_MOUNT_POINT"
    mkdir -p "$LOW_RES_MOUNT_POINT"
    mkdir -p "$CONFIG_MOUNT_POINT"
    gcsfuse "$HIGH_RES_BUCKET" "$HIGH_RES_MOUNT_POINT"
    gcsfuse "$LOW_RES_BUCKET" "$LOW_RES_MOUNT_POINT"
    gcsfuse "$CONFIG_BUCKET" "$CONFIG_MOUNT_POINT"
    ```

1. Next, you need to inform the application where to find the GCS Fuse mount point. This is done by adding the `gcs_fuse_mount_point` setting to your local configuration file (`configs/.env.local.toml`). The following command automates this update. It adds the configuration under the `[storage]` section
```sh
MOUNT_POINT_PATH=$(cd "$HOME/media-search-mnt" && pwd)
sed -i "/low_res_output_bucket/a gcs_fuse_mount_point = \"$MOUNT_POINT_PATH\"" configs/.env.local.toml
```

## Manually trigger pipeline executions
You can manually trigger the `proxy-generation-workflow` and `analyze-workflow `for development and testing, which are otherwise initiated automatically by file uploads.

To execute a workflow, use the corresponding gcloud command below. In each command, replace the `<file-name>` placeholder with the name of the file you wish to process.

* Trigger `proxy-generation-workflow` run:
    ```sh
    gcloud workflows run proxy-generation-workflow \
      --project=$(terraform -chdir=build/terraform output -raw project_id) \
      --location=$(terraform -chdir=build/terraform output -raw cloud_run_region) \
      --data='{"data":{"bucket":"'"$(terraform -chdir=build/terraform output -raw high_res_bucket)"'","name":"<file-name>"}}'
    ```
* Trigger `analyze-workflow` run:
    ```sh
    gcloud workflows run analyze-workflow \
      --project=$(terraform -chdir=build/terraform output -raw project_id) \
      --location=$(terraform -chdir=build/terraform output -raw cloud_run_region) \
      --data='{"data":{"bucket":"'"$(terraform -chdir=build/terraform output -raw low_res_bucket)"'","name":"<file-name>"}}'
    ```

### State Management and Resiliency

A key feature of the `analyze-workflow` is its use of Cloud Storage object metadata to track the state of the analysis process. Intermediate results and step-completion markers are saved as custom metadata on the processed file.

This design provides two main benefits:
1.  **Efficiency**: The pipeline can reuse previously generated information, saving computational resources.
1.  **Resiliency**: If a step in the workflow is interrupted, it can resume from the last successfully completed step by checking the object's metadata, making the entire process more robust and faster on retries.


For testing and debugging, you can inspect the output of individual workflow steps or trigger a re-run by managing the object's custom metadata in Cloud Storage.

To retrieve the output URI for a specific step, run the following command:

```sh
gcloud storage objects describe "gs://<YOUR_LOW_RES_BUCKET>/<FILE_NAME>.mp4" --format="value(custom_fields.<METADATA_KEY>)"
```

And to force a specific step to re-run by removing its corresponding metadata marker from the Cloud Storage object. To remove a metadata key run:
```sh
gcloud storage objects update "gs://<YOUR_LOW_RES_BUCKET>/<FILE>.mp4" --remove-custom-metadata=<THE_STEP_METADATA_KEY>
```
The next time the workflow runs on this file, it will rerun the step as the corresponding completion marker is missing.

The metadata keys are created in the following order during the `analyze-workflow`:
*   `ims_content_length`: Stores the total duration of the video file in seconds.
*   `ims_content_type`: Indicates the classified content type based on prompt configurations.
*   `ims_content_summary`: Contains the AI-generated summary and identified segment timecodes.
*   `ims_content_summary_n_m`: For videos over 20 minutes, stores a summary for a chunk from second `n` to `m`. For instance 0_1200 represents the first 20 minutes of the video file.
*   `ims_segment_summary_x`: Stores the AI-generated description for a specific segment, identified by sequence number `x`.
*   `ims_segment_summary_all`: A marker indicating that descriptions for all segments have been generated.
*   `ims_persist`: A marker indicating that all generated metadata has been saved to BigQuery. Also contains the unique id for the video file generated and stored BigQuery.
*   `ims_generate_embeddings`: A marker indicating that vector embeddings for all segments have been generated.

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
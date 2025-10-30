# Media file processing pipeline
The resources in this folder define an event-driven media processing pipeline orchestrated by Cloud Workflows.

Here's how it works:
1.  **High-Res Upload & Proxy Generation**: Uploading a video to the high-resolution Cloud Storage bucket triggers an Eventarc trigger. This trigger invokes the `proxy-generation-workflow`. The workflow starts a Cloud Run Job that uses FFmpeg to create a lower-resolution proxy of the video and saves it to the low-resolution bucket.
2.  **Low-Res Upload & Media Analysis**: The creation of the proxy file in the low-resolution bucket triggers a second Eventarc trigger. This invokes the `analyze-workflow`. This workflow starts another Cloud Run Job that performs the AI-driven media intelligence analysis using Gemini.

The analysis process produces the following outcomes:
*   **Video Segmentation**: Gemini analyzes the proxy file to identify and provide timecodes for each individual shot.
*   **Contextual Analysis**: Gemini provides text-based context for each shot segment.
*   **Data Persistence**: Timecodes, context, and vector embeddings are persisted to BigQuery.

## Deployment
Use terraform to deploy the required resource for the media file processing pipeline.

1. Create the `terraform.tfvars` file by making a copy of the `terraform.tfvars.example` file.
```sh
cp terraform.tfvars.example terraform.tfvars
```
1. Update terraform.tfvars with your Google Cloud project ID and desired names for the Cloud Storage buckets (for high-resolution and low-resolution media files).
1. Initialize Terraform and apply the configuration to create the resources:
```sh
terraform init
terraform apply
```

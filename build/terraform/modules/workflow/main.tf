module "workflows_service_account" {
  source     = "github.com/terraform-google-modules/terraform-google-service-accounts?ref=a11d4127eab9b51ec9c9afdaf51b902cd2c240d9" #commit hash of version 4.0.0
  project_id = var.project_id
  names      = ["workflows-run-job-sa"]
  project_roles = [
    "${var.project_id}=>roles/eventarc.eventReceiver",
    "${var.project_id}=>roles/logging.logWriter",
    "${var.project_id}=>roles/run.admin",
    "${var.project_id}=>roles/workflows.invoker",
  ]
  display_name = "Media Search Web Service Account"
  description  = "specific custom service account for Web APP"
}

# Grant the Cloud Storage service agent permission to publish Pub/Sub topics
data "google_storage_project_service_account" "gcs_account" {}

resource "google_project_iam_member" "pubsubpublisher" {
  project = var.project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"
}

resource "time_sleep" "wait_for_iam_propagation" {
  create_duration = "60s"
  depends_on      = [google_project_iam_member.pubsubpublisher]
}

resource "google_cloud_run_v2_job" "generate_proxy_job" {
  name     = "generate-proxy-job"
  location = var.region
  project  = var.project_id
  template {
    template {
      volumes {
        name = "high-res-bucket"
        gcs {
          bucket    = var.high_res_bucket
          read_only = true
        }
      }
      volumes {
        name = "low-res-bucket"
        gcs {
          bucket = var.low_res_bucket
        }
      }
      containers {
        image = var.proxy_generator_container
        volume_mounts {
          name       = "high-res-bucket"
          mount_path = "/mnt/${var.high_res_bucket}"
        }
        volume_mounts {
          name       = "low-res-bucket"
          mount_path = "/mnt/${var.low_res_bucket}"
        }
        env {
          name  = "OUTPUT_FOLDER"
          value = var.low_res_bucket
        }
        resources {
          limits = {
            cpu    = "8"
            memory = "16Gi"
          }
        }
      }
      service_account = var.media_search_service_account_email
      timeout         = "3600s"
    }
  }
}

resource "google_workflows_workflow" "proxy_workflow" {
  name            = "proxy-generation-workflow"
  region          = var.region
  description     = "Workflow to trigger Cloud Run Job for proxy generation"
  service_account = module.workflows_service_account.email


  source_contents = <<EOF
  main:
    params: [event]
    steps:
      - init:
          assign:
            - input_file: $${event.data.bucket + "/" + event.data.name}
      - run_job:
          call: googleapis.run.v1.namespaces.jobs.run
          args:
              connector_params:
                  skip_polling: True
              name: "namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.generate_proxy_job.name}"
              location: ${var.region}
              body:
                  overrides:
                      containerOverrides:
                          env:
                              - name: "INPUT_FILE"
                                value: $${input_file}
          result: job_execution
      - finish:
          return: $${job_execution}
  EOF
}

resource "google_eventarc_trigger" "proxy_workflow_trigger" {
  name     = "proxy-workflow-trigger"
  location = var.region
  project  = var.project_id

  matching_criteria {
    attribute = "type"
    value     = "google.cloud.storage.object.v1.finalized"
  }

  matching_criteria {
    attribute = "bucket"
    value     = var.high_res_bucket
  }

  service_account = module.workflows_service_account.email

  destination {
    workflow = google_workflows_workflow.proxy_workflow.id
  }
  depends_on = [time_sleep.wait_for_iam_propagation]
}

resource "google_cloud_run_v2_job" "media_analysis_job" {
  name     = "media-analysis-job"
  location = var.region
  project  = var.project_id
  template {
    template {
      volumes {
        name = "low-res-bucket"
        gcs {
          bucket = var.low_res_bucket
        }
      }
      volumes {
        name = "config-bucket"
        gcs {
          bucket = var.config_bucket
        }
      }
      containers {
        image = var.media_analysis_container
        volume_mounts {
          name       = "low-res-bucket"
          mount_path = "/mnt/${var.low_res_bucket}"
        }
        volume_mounts {
          name       = "config-bucket"
          mount_path = "/mnt/${var.config_bucket}"
        }
        env {
          name  = "GCP_CONFIG_PREFIX"
          value = "/mnt/${var.config_bucket}"
        }
        resources {
          limits = {
            cpu    = "8"
            memory = "4Gi"
          }
        }
      }
      service_account = var.media_search_service_account_email
      timeout         = "3600s"
    }
  }
}

resource "google_workflows_workflow" "analyze_workflow" {
  name            = "analyze-workflow"
  region          = var.region
  description     = "Workflow to trigger Cloud Run Job for analyze the media content"
  service_account = module.workflows_service_account.email


  source_contents = <<EOF
  main:
    params: [event]
    steps:
      - init:
          assign:
            - input_file: $${event.data.bucket + "/" + event.data.name}
      - run_job:
          call: googleapis.run.v1.namespaces.jobs.run
          args:
              connector_params:
                  skip_polling: True
              name: "namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.media_analysis_job.name}"
              location: ${var.region}
              body:
                  overrides:
                      containerOverrides:
                          env:
                              - name: "INPUT_FILE"
                                value: $${input_file}
          result: job_execution
      - finish:
          return: $${job_execution}
  EOF
}

resource "google_eventarc_trigger" "analyze_workflow_trigger" {
  name     = "analyze-workflow-trigger"
  location = var.region
  project  = var.project_id

  matching_criteria {
    attribute = "type"
    value     = "google.cloud.storage.object.v1.finalized"
  }

  matching_criteria {
    attribute = "bucket"
    value     = var.low_res_bucket
  }

  service_account = module.workflows_service_account.email

  destination {
    workflow = google_workflows_workflow.analyze_workflow.id
  }
  depends_on = [time_sleep.wait_for_iam_propagation]
}
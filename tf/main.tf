# Terraform for this Fido fetch project on GCP.
#
# Default shape:
#   - one GCS bucket for the output
#   - one Cloud Run job (built from cloudbuild.yaml + Dockerfile)
#   - one Cloud Scheduler trigger that runs the job on a cadence
#   - service accounts + IAM bindings for the job and the scheduler
#
# Rename the resources (bucket name, job name, scheduler name, SA ids)
# to match the project name from DESIGN.md before the first `terraform
# apply`.

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0.0"
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
}

variable "project" {
  description = "GCP project to deploy into."
  type        = string
}

variable "region" {
  description = "GCP region for Cloud Run + GCS."
  type        = string
  default     = "us-central1"
}

variable "image_tag" {
  description = "Image tag for the Cloud Run job image (typically SHORT_SHA)."
  type        = string
}

variable "source_url" {
  description = "Upstream source URL or URI passed to --source-url."
  type        = string
}

variable "schedule" {
  description = "Cloud Scheduler cron expression that triggers the job."
  type        = string
  default     = "0 * * * *"
}

# --- Storage -----------------------------------------------------------------

resource "google_storage_bucket" "output" {
  name                        = "${var.project}-fido-fetch-output"
  location                    = var.region
  force_destroy               = false
  uniform_bucket_level_access = true
}

# --- Cloud Run job -----------------------------------------------------------

resource "google_service_account" "job" {
  account_id   = "fido-fetch-job"
  display_name = "Fido fetch Cloud Run job"
}

resource "google_storage_bucket_iam_member" "job_writer" {
  bucket = google_storage_bucket.output.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.job.email}"
}

resource "google_cloud_run_v2_job" "fetch" {
  name     = "fido-fetch"
  location = var.region

  template {
    template {
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project}/fido-fetch/fido-fetch:${var.image_tag}"
        args = [
          "--source-url=${var.source_url}",
          "--output=gs://${google_storage_bucket.output.name}",
        ]
      }
      service_account = google_service_account.job.email
      timeout         = "3600s"
    }
  }
}

# --- Cloud Scheduler: trigger the Cloud Run job ------------------------------

resource "google_service_account" "scheduler" {
  account_id   = "fido-fetch-scheduler"
  display_name = "Fido fetch Cloud Scheduler"
}

resource "google_cloud_scheduler_job" "fetch" {
  name     = "fido-fetch"
  region   = var.region
  schedule = var.schedule

  http_target {
    http_method = "POST"
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project}/jobs/${google_cloud_run_v2_job.fetch.name}:run"
    oauth_token {
      service_account_email = google_service_account.scheduler.email
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
}

terraform {
  backend "gcs" {
    bucket = "ossf-feeds-tf-state"
    prefix = "terraform/state"
  }
}

locals {
  services = [
    "cloudbuild.googleapis.com",
    "run.googleapis.com",
    "cloudscheduler.googleapis.com",
    "cloudresourcemanager.googleapis.com",
  ]
}

resource "google_service_account" "run-invoker-account" {
  account_id   = "run-invoker-sa"
  display_name = "Feed Run Invoker"
}

resource "google_project_iam_member" "run-invoker-iam" {
  role   = "roles/run.invoker"
  member = "serviceAccount:${google_service_account.run-invoker-account.email}"
}

resource "google_project_service" "services" {
  for_each           = toset(local.services)
  service            = each.value
  disable_on_destroy = false
}

resource "google_pubsub_topic" "feed-topic" {
  name = "feed-topic"
}

resource "google_storage_bucket" "feed-functions-bucket" {
  name          = "${var.project}-feed-functions-bucket"
  force_destroy = true
}

module "pypi_scheduler" {
  source = "./scheduler"

  pkg-ecosystem         = "pypi"
  project               = var.project
  region                = var.region
  service-account-email = google_service_account.run-invoker-account.email
  pubsub-topic-feed-id  = google_pubsub_topic.feed-topic.id
}

module "npm_scheduler" {
  source = "./scheduler"

  pkg-ecosystem         = "npm"
  project               = var.project
  region                = var.region
  service-account-email = google_service_account.run-invoker-account.email
  pubsub-topic-feed-id  = google_pubsub_topic.feed-topic.id
}

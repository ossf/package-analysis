provider "google" {
  project = var.project
  region  = var.region
}

terraform {
  backend "gcs" {
    bucket = "ossf-analysis-tf-state"
    prefix = "terraform/state"
  }
}

module "docker_registry" {
  source = "./docker_registry"

  project = var.project
}

module "build" {
  source = "./build"

  project = var.project
  github_owner = var.github_owner
  github_repo = var.github_repo
}

module "metrics" {
  source = "./metrics"

  project = var.project
}
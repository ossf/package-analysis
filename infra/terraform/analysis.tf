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

  project               = var.project
}

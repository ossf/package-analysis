resource "google_artifact_registry_repository" "gcr_docker" {
  provider = google-beta

  project = var.project
  location = "us"
  repository_id = "gcr.io"
  description = "gcr.io docker container registry for OSSF Malware Analysis Images"
  format = "DOCKER"
}

resource "google_artifact_registry_repository" "us_gcr_docker" {
  provider = google-beta

  project = var.project
  location = "us"
  repository_id = "us.gcr.io"
  description = "us.gcr.io docker container registry for OSSF Malware Analysis Images"
  format = "DOCKER"
}

resource "google_artifact_registry_repository_iam_policy" "policy" {
  provider = google-beta

  project = google_artifact_registry_repository.gcr_docker.project
  location = google_artifact_registry_repository.gcr_docker.location
  repository = google_artifact_registry_repository.gcr_docker.name
  policy_data = data.google_iam_policy.public_registry_policy.policy_data
}

data "google_iam_policy" "public_registry_policy" {
  binding {
    role = "roles/artifactregistry.reader"

    members = [
      "allUsers",
    ]
  }
}
# Google Cloud Build Triggers

resource "google_cloudbuild_trigger" "image-build-trigger" {
  name = "image-build-trigger"
  project = var.project

  github {
    owner = var.github_owner
    name = var.github_repo
    push {
        tag = "^rel-[0-9]+$"
    }
  }

  filename = "build/cloudbuild.yaml"
}

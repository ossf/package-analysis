resource "google_cloud_scheduler_job" "trigger-pypi-scheduler" {
  name        = "trigger-pypi-scheduler"
  description = "The scheduler that triggers fetching new PyPI packages"
  schedule    = "*/5 * * * *"

  http_target {
    http_method = "POST"
    uri         = google_cloud_run_service.run-pypi.status[0].url

    oidc_token {
      service_account_email = google_service_account.run-invoker-account.email
    }
  }
}

resource "google_cloud_run_service" "run-pypi" {
  name     = "pypi-run-srv"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project}/feeds-pypi"
        env {
          name  = "OSSMALWARE_TOPIC_URL"
          value = "gcppubsub://${google_pubsub_topic.feed-topic.id}"
        }
      }
    }
  }
}

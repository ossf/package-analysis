resource "google_cloud_scheduler_job" "trigger-ecosystem-scheduler" {
  name        = "trigger-${var.pkg-ecosystem}-scheduler"
  description = "Scheduler for fetching new ${var.pkg-ecosystem} packages"
  schedule    = "*/5 * * * *"

  http_target {
    http_method = "POST"
    uri         = google_cloud_run_service.run-scheduler.status[0].url

    oidc_token {
      service_account_email = var.service-account-email
    }
  }
}

resource "google_cloud_run_service" "run-scheduler" {
  name     = "${var.pkg-ecosystem}-run-srv"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project}/feeds-${var.pkg-ecosystem}"
        env {
          name  = "OSSMALWARE_TOPIC_URL"
          value = "gcppubsub://${var.pubsub-topic-feed-id}"
        }
      }
    }
  }
}

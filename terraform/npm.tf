resource "google_cloud_scheduler_job" "trigger-npm-scheduler" {
  name        = "trigger-npm-scheduler"
  description = "The scheduler that triggers fetching new npm packages"
  schedule    = "*/5 * * * *"

  http_target {
    http_method = "POST"
    uri         = google_cloud_run_service.run-npm.status[0].url

    oidc_token {
      service_account_email = google_service_account.run-invoker-account.email
    }
  }
}

resource "google_cloud_run_service" "run-npm" {
  name     = "npm-run-srv"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project}/feeds-npm"
        env {
          name  = "OSSMALWARE_TOPIC_URL"
          value = "gcppubsub://${google_pubsub_topic.feed-topic.id}"
        }
      }
    }
  }
}

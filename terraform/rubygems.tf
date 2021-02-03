resource "google_cloud_scheduler_job" "trigger-rubygems-scheduler" {
  name        = "trigger-rubygems-scheduler"
  description = "The scheduler that triggers fetching new RubyGems packages"
  schedule    = "*/5 * * * *"

  http_target {
    http_method = "POST"
    uri         = google_cloud_run_service.run-rubygems.status[0].url

    oidc_token {
      service_account_email = google_service_account.run-invoker-account.email
    }
  }
}

resource "google_cloud_run_service" "run-rubygems" {
  name     = "rubygems-run-srv"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project}/feeds-rubygems"
        env {
          name  = "OSSMALWARE_TOPIC_URL"
          value = "gcppubsub://${google_pubsub_topic.feed-topic.id}"
        }
      }
    }
  }
}

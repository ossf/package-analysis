resource "google_logging_metric" "analysis_requests_metric" {
  name   = "analysis/request_count"
  filter = <<-EOT
    resource.type="k8s_container"
    resource.labels.project_id="ossf-malware-analysis"
    resource.labels.cluster_name="analysis-cluster"
    resource.labels.namespace_name="default"
    labels.k8s-pod/app="workers"
    "Got request"
  EOT
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    labels {
      key         = "ecosystem"
      value_type  = "STRING"
      description = "package ecosystem"
    }
  }
  label_extractors = {
    "ecosystem" = "EXTRACT(labels.ecosystem)"
  }
  project = var.project
}

resource "google_logging_metric" "analysis_success_metric" {
  name   = "analysis/success_count"
  filter = <<-EOT
    resource.type="k8s_container"
    resource.labels.project_id="ossf-malware-analysis"
    resource.labels.cluster_name="analysis-cluster"
    resource.labels.namespace_name="default"
    labels.k8s-pod/app="workers"
    "Analysis completed sucessfully"
  EOT
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    labels {
      key         = "ecosystem"
      value_type  = "STRING"
      description = "package ecosystem"
    }
  }
  label_extractors = {
    "ecosystem" = "EXTRACT(labels.ecosystem)"
  }
  project = var.project
}

resource "google_logging_metric" "analysis_run_error_metric" {
  name   = "analysis/run_error_count"
  filter = <<-EOT
    resource.type="k8s_container"
    resource.labels.project_id="ossf-malware-analysis"
    resource.labels.cluster_name="analysis-cluster"
    resource.labels.namespace_name="default"
    labels.k8s-pod/app="workers"
    ("Analysis run failed" OR "Analysis error - other")
  EOT
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    labels {
      key         = "ecosystem"
      value_type  = "STRING"
      description = "package ecosystem"
    }
  }
  label_extractors = {
    "ecosystem" = "EXTRACT(labels.ecosystem)"
  }
  project = var.project
}

resource "google_logging_metric" "analysis_error_metric" {
  name   = "analysis/error_count"
  filter = <<-EOT
    resource.type="k8s_container"
    resource.labels.project_id="ossf-malware-analysis"
    resource.labels.cluster_name="analysis-cluster"
    resource.labels.namespace_name="default"
    labels.k8s-pod/app="workers"
    "Analysis error - analysis"
  EOT
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    labels {
      key         = "ecosystem"
      value_type  = "STRING"
      description = "package ecosystem"
    }
  }
  label_extractors = {
    "ecosystem" = "EXTRACT(labels.ecosystem)"
  }
  project = var.project
}

resource "google_logging_metric" "analysis_timeout_metric" {
  name   = "analysis/timeout_count"
  filter = <<-EOT
    resource.type="k8s_container"
    resource.labels.project_id="ossf-malware-analysis"
    resource.labels.cluster_name="analysis-cluster"
    resource.labels.namespace_name="default"
    labels.k8s-pod/app="workers"
    "Analysis error - timeout"
  EOT
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    labels {
      key         = "ecosystem"
      value_type  = "STRING"
      description = "package ecosystem"
    }
  }
  label_extractors = {
    "ecosystem" = "EXTRACT(labels.ecosystem)"
  }
  project = var.project
}

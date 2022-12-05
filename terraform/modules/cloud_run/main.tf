# This module assumes the 'vpcaccess.googleapis.com' service has already been
# enabled elsewhere.
resource "google_vpc_access_connector" "svc" {
  provider = google-beta

  name          = "svc-conn"
  region        = var.region
  ip_cidr_range = "10.8.0.0/28"
  network       = var.vpc_name
  project       = var.project_id
}

locals {
  # See https://cloud.google.com/iam/docs/service-agents
  cloud_run_service_agent = "service-${var.project_number}@serverless-robot-prod.iam.gserviceaccount.com"
}

resource "google_kms_key_ring_iam_member" "secrets_access" {
  key_ring_id = var.sops_keyring_id
  role        = "roles/cloudkms.cryptoKeyDecrypter"
  member      = "serviceAccount:${var.service_account_email}"
}

resource "google_artifact_registry_repository_iam_member" "repo_access" {
  provider = google-beta

  location   = var.service_image.region
  repository = var.service_image.repo
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${local.cloud_run_service_agent}"
  project    = var.service_image.project_id
}

locals {
  service_image = "${var.service_image.region}-docker.pkg.dev/${var.service_image.project_id}/${var.service_image.repo}/${var.service_image.image_name}:latest"
}

resource "google_cloud_run_service" "default" {
  name     = var.service_name
  location = var.region
  project  = var.project_id

  template {
    spec {
      service_account_name = var.service_account_email
      containers {
        image = local.service_image
        env {
          name  = "CONFIG"
          value = var.config_path
        }
      }
    }

    metadata {
      annotations = {
        # Limit scale up to prevent any cost blow outs!
        "autoscaling.knative.dev/maxScale" = "5"
        # Use the VPC Connector
        "run.googleapis.com/vpc-access-connector" = google_vpc_access_connector.svc.id
        # Only DB trafic from the service should go through the VPC Connector
        "run.googleapis.com/vpc-access-egress" = "private-ranges-only"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [
    # We can't pull from the registry until we have access to it.
    google_artifact_registry_repository_iam_member.repo_access
  ]

  lifecycle {
    # We ignore these because they change every deploy and we don't care about
    # them.
    ignore_changes = [
      template.0.metadata.0.annotations["client.knative.dev/user-image"],
      template.0.metadata.0.annotations["run.googleapis.com/client-name"],
      template.0.metadata.0.annotations["run.googleapis.com/client-version"],
    ]
  }
}

resource "google_project_iam_member" "service_account" {
  for_each = toset([
    # Both of these are required for creating session cookies.
    "roles/serviceusage.serviceUsageConsumer",
    "roles/firebaseauth.admin",
  ])

  project = var.project_id
  role    = each.key
  member  = "serviceAccount:${var.service_account_email}"
}

# We need to give the service account permission to sign tokens on behalf
# of...itself. Without this, issuing session cookies will fail.
resource "google_service_account_iam_member" "service_account" {
  for_each = toset([
    # Both of these are required for creating session cookies.
    "roles/iam.serviceAccountTokenCreator",
    "roles/iam.serviceAccountUser",
  ])

  service_account_id = var.service_account_name
  role               = each.key
  member             = "serviceAccount:${var.service_account_email}"
}

provider "google" {
  region = var.region
}

provider "google-beta" {
  region = var.region
}

module "project" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 14.1"

  billing_account          = var.billing_account
  name                     = "${var.project_id_prefix}-shared"
  random_project_id        = true
  random_project_id_length = 4

  activate_apis = [
    "artifactregistry.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "serviceusage.googleapis.com",
    "iam.googleapis.com",
    "cloudkms.googleapis.com",
    "firebase.googleapis.com",
    "cloudbilling.googleapis.com",
    "servicenetworking.googleapis.com",
    "sqladmin.googleapis.com",
  ]

  # Only specify the org_id if the folder wasn't specified.
  org_id    = var.folder_id == "" ? var.org_id : ""
  folder_id = var.folder_id

  default_service_account = "deprivilege"
}

resource "google_kms_key_ring" "sops" {
  name     = "sops"
  location = "global"
  project  = module.project.project_id

  depends_on = [
    module.project.enabled_apis,
  ]
}

# Key for encrypting/decrypting app secrets with https://github.com/mozilla/sops
resource "google_kms_crypto_key" "developers" {
  name     = "developers"
  key_ring = google_kms_key_ring.sops.id

  # Create a new key version every 90 days.
  rotation_period = "7776000s"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_artifact_registry_repository" "registry" {
  provider = google-beta

  location      = var.region
  repository_id = var.app_name
  description   = "Docker images for the ${var.app_name} service"
  format        = "DOCKER"
  project       = module.project.project_id

  depends_on = [
    module.project.enabled_apis,
  ]
}

resource "google_service_account" "terraform" {
  account_id   = "terraform"
  display_name = "Terraform service account for ${var.app_name}"
  description  = "Service account that applies Terraform configuration for ${var.app_name} projects"
  project      = module.project.project_id
}

# See [1] for where these roles come from. Copied here:
# In order to execute this module you must have a Service Account with the following roles:
#     - roles/resourcemanager.folderViewer on the folder that you want to create the project in
#     - roles/resourcemanager.organizationViewer on the organization
#     - roles/resourcemanager.projectCreator on the organization
#     - roles/billing.user on the organization
# [1] https://github.com/terraform-google-modules/terraform-google-project-factory

resource "google_folder_iam_member" "terraform" {
  count = var.folder_id == "" ? 0 : 1

  folder = "folders/${var.folder_id}"
  role   = "roles/resourcemanager.folderViewer"
  member = "serviceAccount:${google_service_account.terraform.email}"
}

resource "google_organization_iam_member" "terraform" {
  for_each = toset([
    "roles/resourcemanager.organizationViewer",
    "roles/resourcemanager.projectCreator",
  ])

  org_id = var.org_id
  role   = each.key
  member = "serviceAccount:${google_service_account.terraform.email}"
}

resource "google_billing_account_iam_member" "terraform" {
  billing_account_id = var.billing_account
  role               = "roles/billing.user"
  member             = "serviceAccount:${google_service_account.terraform.email}"
}

resource "google_project_iam_member" "owner" {
  project = module.project.project_id
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.terraform.email}"
}

data "google_client_openid_userinfo" "current_user" {}

resource "google_service_account_iam_member" "current_user_as_tf" {
  for_each = toset([
    "roles/iam.serviceAccountUser",
    "roles/iam.serviceAccountTokenCreator",
  ])
  service_account_id = google_service_account.terraform.name
  role               = each.key
  member             = "user:${data.google_client_openid_userinfo.current_user.email}"
}

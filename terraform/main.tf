data "terraform_remote_state" "shared" {
  backend = "local"

  config = {
    path = "${path.module}/shared/terraform.tfstate"
  }
}

provider "google" {
  region                      = var.region
  impersonate_service_account = data.terraform_remote_state.shared.outputs.terraform_service_account
}

provider "google-beta" {
  region                      = var.region
  impersonate_service_account = data.terraform_remote_state.shared.outputs.terraform_service_account
}

locals {
  # Dev and prod workspaces will have mostly identical resources, but we
  # don't need all of the same resources in the local environment, like a Cloud
  # SQL database for example.
  is_local = terraform.workspace == "local"

  config_path = "/configs/${terraform.workspace}.conf"

  # Services that are only required by non-local resources should go here.
  non_local_services = [
    # Used to deploy our frontend to Firebase Hosting + Functions.
    "cloudbuild.googleapis.com",
    "cloudfunctions.googleapis.com",
    # Used in the database module.
    "servicenetworking.googleapis.com",
    # Needed to access the database through Cloud Shell/gcloud.
    "sqladmin.googleapis.com",
    # Used by Cloud Run to access the DB.
    "vpcaccess.googleapis.com",
    # Used by Cloud Run because, well, it is Cloud Run.
    "run.googleapis.com",
  ]

  # We specify project_services in project creation instead of with the
  # modules/resources they're used by because we don't want multiple resources
  # trying to turn on/off the same service.
  project_services = concat([
    "cloudkms.googleapis.com",
    "identitytoolkit.googleapis.com", # See the auth module.
    "firebase.googleapis.com",        # See the auth module.
    "storage.googleapis.com",         # See the assets module.
  ], local.is_local ? [] : local.non_local_services)
}

module "project" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 14.1"

  billing_account          = var.billing_account
  name                     = "${var.project_id_prefix}-${var.env}"
  random_project_id        = true
  random_project_id_length = 4

  activate_apis = local.project_services

  # Only specify the org_id if the folder wasn't specified.
  org_id    = var.folder_id == "" ? var.org_id : ""
  folder_id = var.folder_id
}

resource "google_firebase_project" "default" {
  provider = google-beta
  project  = module.project.project_id
}

# The default VPC network
resource "google_compute_network" "default" {
  provider = google-beta

  name                    = "default"
  description             = "Default network for the project"
  auto_create_subnetworks = true
  project                 = module.project.project_id

  depends_on = [
    module.project.enabled_apis,
  ]
}

resource "google_project_iam_audit_config" "default" {
  project = module.project.project_id
  service = "allServices"

  audit_log_config {
    log_type = "ADMIN_READ"
  }
  audit_log_config {
    log_type = "DATA_READ"
  }
  audit_log_config {
    log_type = "DATA_WRITE"
  }
}

resource "google_kms_key_ring" "app" {
  name     = var.app_name
  location = "global"
  project  = module.project.project_id

  depends_on = [
    module.project.enabled_apis,
  ]
}

# Key for encrypting/decrypting app secrets with https://github.com/mozilla/sops
resource "google_kms_crypto_key" "sops" {
  name     = "sops"
  key_ring = google_kms_key_ring.app.id

  # Create a new key version every 90 days.
  rotation_period = "7776000s"

  lifecycle {
    prevent_destroy = true
  }
}

module "database" {
  # Our local environment doesn't need a Cloud SQL database.
  count = local.is_local ? 0 : 1

  source = "./modules/database"

  db_name    = var.app_name
  region     = var.region
  db_tier    = var.db_tier
  vpc_id     = google_compute_network.default.id
  project_id = module.project.project_id
  env        = var.env

  depends_on = [
    module.project.enabled_apis,
  ]
}

resource "google_service_account" "app" {
  account_id   = var.app_name
  display_name = "Cloud Run Service Account for ${var.app_name}"
  description  = "Service account that the ${var.app_name} service runs as"
  project      = module.project.project_id
}

module "cloud_run" {
  # Our local environment doesn't need a Cloud Run deployment.
  count = local.is_local ? 0 : 1

  source = "./modules/cloud_run"

  service_account_name  = google_service_account.app.name
  service_account_email = google_service_account.app.email

  region         = var.region
  service_name   = var.app_name
  vpc_id         = google_compute_network.default.id
  vpc_name       = google_compute_network.default.name
  project_id     = module.project.project_id
  project_number = module.project.project_number
  env            = var.env

  service_image = {
    project_id = data.terraform_remote_state.shared.outputs.project_id
    region     = var.region
    repo       = var.app_name
    image_name = "server"
  }
  config_path = local.config_path

  sops_keyring_id = google_kms_key_ring.app.id

  depends_on = [
    module.project.enabled_apis,
  ]
}

module "auth" {
  source = "./modules/auth"

  display_name = "${var.app_name} ${title(var.env)}"
  project_id   = module.project.project_id

  depends_on = [
    google_firebase_project.default
  ]
}

# See https://cloud.google.com/sql/docs/postgres/configure-private-services-access
resource "google_compute_global_address" "main_db_ip" {
  provider = google-beta

  name          = "main-db-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 20
  network       = var.vpc_id
  project       = var.project_id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  provider = google-beta

  network                 = var.vpc_id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.main_db_ip.name]
}

resource "random_id" "db_name_suffix" {
  byte_length = 4
}

resource "google_sql_database_instance" "default" {
  provider = google-beta

  name             = "${var.db_name}-${random_id.db_name_suffix.hex}"
  region           = var.region
  database_version = "POSTGRES_14"
  project          = var.project_id

  depends_on = [google_service_networking_connection.private_vpc_connection]

  settings {
    tier = var.db_tier

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.vpc_id
    }
  }
}

resource "google_sql_database" "default" {
  name     = replace(var.db_name, "-", "_")
  instance = google_sql_database_instance.default.name
  project  = var.project_id
}

resource "random_password" "password" {
  length  = 24
  special = false
}

resource "google_sql_user" "users" {
  name     = "postgres"
  instance = google_sql_database_instance.default.name
  password = random_password.password.result
  project  = var.project_id
}

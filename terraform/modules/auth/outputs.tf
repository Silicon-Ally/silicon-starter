# Based heavily on [1], though we just produce these as outputs instead of
# copying them to a GCS bucket.
# [1] https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/firebase_web_app

output "app_id" {
  value = google_firebase_web_app.default.app_id
}

output "api_key" {
  value = data.google_firebase_web_app_config.default.api_key
}

output "auth_domain" {
  value = data.google_firebase_web_app_config.default.auth_domain
}

output "storage_bucket" {
  value = lookup(data.google_firebase_web_app_config.default, "storage_bucket", "")
}

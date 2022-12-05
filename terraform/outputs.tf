output "project_id" {
  value = module.project.project_id
}

output "firebase_app_id" {
  value = module.auth.app_id
}

output "firebase_api_key" {
  value = module.auth.api_key
}

output "firebase_auth_domain" {
  value = module.auth.auth_domain
}

output "firebase_storage_bucket" {
  value = module.auth.storage_bucket
}

output "postgres_password" {
  value     = module.database.*.postgres_password
  sensitive = true
}


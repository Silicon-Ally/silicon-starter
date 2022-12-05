output "project_id" {
  description = "GCP Project ID for the shared project"
  value       = module.project.project_id
}

output "terraform_service_account" {
  description = "The service account to use when applying Terraform for other projects"
  value       = google_service_account.terraform.email
}

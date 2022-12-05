output "instance_name" {
  description = "The name of the created database instance."
  value       = google_sql_database_instance.default.name
}

output "postgres_password" {
  description = "Password for the 'postgres' user"
  value       = google_sql_user.users.password
}

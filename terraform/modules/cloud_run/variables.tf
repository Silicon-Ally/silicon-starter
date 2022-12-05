variable "region" {
  description = "The region to create Cloud Run resources in."
}

variable "service_name" {
  description = "Name to use for the Cloud Run service."
  type        = string
  validation {
    condition     = can(regex("^[a-z-]+$", var.service_name))
    error_message = "The service_name should be a snake-case identifier"
  }
}

variable "service_account_name" {
  description = "The full name/ID of the service account to run as."
}

variable "service_account_email" {
  description = "The email of the service account to run as."
}

variable "vpc_id" {
  description = "The ID of the VPC network to attach this service to."
}

variable "vpc_name" {
  description = "The name of the VPC network to attach this service to."
}

variable "env" {
  description = "The environment this resource is being created in, used as a resource suffix."
}

variable "project_id" {
  description = "The project to create Cloud Run resources in."
}

variable "project_number" {
  description = "The number of the project to create Cloud Run resources in."
}

variable "service_image" {
  description = "The URL of the Docker image for this service."
  type = object({
    project_id = string
    region     = string
    repo       = string
    image_name = string
  })
}

variable "sops_keyring_id" {
  description = "The ID of the keyring that we should use to encrypt/decrypt."
}

variable "config_path" {
  description = "The path of the configuration file to load at runtime."
}

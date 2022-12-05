variable "billing_account" {
  description = "The billing account to link to the project."
}

variable "project_id_prefix" {
  description = "The prefix to add to all created projects. Project names are $PREFIX-$ENV-$RANDOM_SUFFIX"
}

variable "app_name" {
  description = "Name of this application, goes into descriptions and resource names."
  type        = string
  validation {
    condition     = can(regex("^[a-z-]+$", var.app_name))
    error_message = "The app_name should be a snake-case identifier"
  }
}

variable "region" {
  description = "The default region to create resources in."
}

variable "env" {
  description = "The environment to create, gets incorporated into the project name/ID"
}

variable "folder_id" {
  description = "The ID of the GCP folder to create projects in, or blank for root."
}

variable "org_id" {
  description = "The ID of the Google Cloud org to create these projects in. Only one of this or folder_id can be specified."
  validation {
    condition     = var.org_id != ""
    error_message = "The org_id must be specified"
  }
}

variable "db_tier" {
  description = "The size of the database to create."
}

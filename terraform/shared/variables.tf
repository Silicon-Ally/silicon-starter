variable "billing_account" {
  description = "The billing account to link to the project."
}

variable "project_id_prefix" {
  description = "The prefix to add to all created projects. Project names are $PREFIX-$ENV-$RANDOM_SUFFIX"
}

variable "app_name" {
  description = "Name of this application, goes into descriptions and resource names."
}

variable "region" {
  description = "The default region to create resources in."
}

variable "folder_id" {
  description = "The ID of the GCP folder for this client."
}

variable "org_id" {
  description = "The ID of the Google Cloud org to create these projects in. Only one of this or folder_id can be specified."
}


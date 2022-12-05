variable "region" {
  description = "The region to create database resources in."
}

variable "db_tier" {
  description = "The size of the database to create."
}

variable "vpc_id" {
  description = "The ID of the VPC network to create this database in."
}

variable "env" {
  description = "The environment this resource is being created in, used as a resource suffix."
}

variable "project_id" {
  description = "The project to create database resources in."
}

variable "db_name" {
  description = "Name to use for both the Cloud SQL instance and for the database itself within the instance."
  type        = string
  validation {
    condition     = can(regex("^[a-z-]+$", var.db_name))
    error_message = "The db_name should be a snake-case identifier"
  }
}

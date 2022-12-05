terraform {
  backend "local" {
    workspace_dir = "tfstate/"
  }

  required_version = "~> 1.3.6"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.44"
    }

    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 4.44"
    }
  }
}

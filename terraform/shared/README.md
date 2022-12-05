# Shared Terraform Configuration

This directory contains Terraform configuration that
doesn't belong to an environment (e.g. `dev` or `prod`).
These are shared resources, like Docker images, that
shouldn't be associated directly with an environment.

# Usage

First, initialize your Terraform state with:

```bash
cd terraform/shared
terraform init -var-file "../terraform.tfvars"
```

This should only need to be run once per checkout.

```bash
# To view the pending changes.
terraform plan -var-file "../terraform.tfvars"

# To apply the pending changes.
terraform apply -var-file "../terraform.tfvars"
```

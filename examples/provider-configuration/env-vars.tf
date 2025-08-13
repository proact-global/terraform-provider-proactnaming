# Copyright (c) HashiCorp, Inc.

# Environment Variables Configuration Example
# RECOMMENDED: This method keeps sensitive values out of your Terraform code

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

# Provider configuration using environment variables
# Set these before running Terraform:
#   export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
#   export PROACTNAMING_APIKEY="your-api-key"
#   export PROACTNAMING_ADMIN_PASSWORD="your-admin-password"  # Optional

provider "proactnaming" {
  # All configuration comes from environment variables
  # No explicit configuration needed here
}

# Simple example resource
resource "proactnaming_generate_name" "example" {
  organization  = "test"
  resource_type = "rg"
  application   = "demo"
  function      = "api"
  instance      = "001"
  location      = "euw"
  environment   = "dev"
}

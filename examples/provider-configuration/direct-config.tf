# Direct Configuration Example
# WARNING: This method exposes sensitive values in your Terraform code
# Only use this for testing or when values are managed securely

terraform {
  required_providers {
    proactnaming = {
      source = "hashicorp.com/edu/proactnaming"
    }
  }
}

provider "proactnaming" {
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = "your-api-key-here"
  admin_password = "your-admin-password-here" # Optional
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

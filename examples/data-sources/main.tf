# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

# Configure the provider with your Azure Naming Tool details
provider "proactnaming" {
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = "your-api-key-here"
  admin_password = "your-admin-password" # Optional, only for delete operations
}

# Get all available resource types from the Azure Naming Tool
data "proactnaming_resource_types" "all" {}

# Get a specific generated name
data "proactnaming_generated_name" "existing" {
  id = 123 # Replace with an actual generated name ID
}

# Output available resource types
output "available_resource_types" {
  description = "List of all available Azure resource types"
  value       = data.proactnaming_resource_types.all.resource_types
}

# Output the retrieved generated name
output "existing_name_details" {
  description = "Details of the existing generated name"
  value = {
    name         = data.proactnaming_generated_name.existing.resource_name
    organization = data.proactnaming_generated_name.existing.organization
    environment  = data.proactnaming_generated_name.existing.environment
    application  = data.proactnaming_generated_name.existing.application
  }
  sensitive = false
}

# Example: Use resource types data to validate configuration
locals {
  valid_resource_types = [for rt in data.proactnaming_resource_types.all.resource_types : rt.short_name]

  # Validate that our desired resource type exists
  resource_type_valid = contains(local.valid_resource_types, "rg")
}

# Generate a name using validated resource type
resource "proactnaming_generate_name" "validated_example" {
  count = local.resource_type_valid ? 1 : 0

  organization  = "myorg"
  resource_type = "rg"
  application   = "webapp"
  function      = "core"
  instance      = "001"
  location      = "euw"
  environment   = "dev"
}

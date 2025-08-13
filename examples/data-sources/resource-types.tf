# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

provider "proactnaming" {
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = "your-api-key"
  admin_password = "your-admin-password" # Optional
}

# Query all available resource types
data "proactnaming_resourcetypes" "all" {}

# Example: Use data source results to validate a resource type exists
locals {
  # Extract all short names from the resource types
  available_types = [
    for rt in data.proactnaming_resourcetypes.all.resource_types : rt.short_name
  ]

  # Check if specific types we want to use are available
  required_types = ["rg", "st", "vm", "app", "sql"]

  # Find any missing types
  missing_types = [
    for type in local.required_types : type
    if !contains(local.available_types, type)
  ]
}

# Generate names only if required types are available
resource "proactnaming_generate_name" "conditional_rg" {
  count = length(local.missing_types) == 0 ? 1 : 0

  organization  = "myorg"
  resource_type = "rg"
  application   = "webapp"
  function      = "infra"
  instance      = "001"
  location      = "euw"
  environment   = "dev"
}

# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

provider "proactnaming" {
  # Uses PROACTNAMING_HOST and PROACTNAMING_APIKEY environment variables,
  # or set them explicitly:
  # host   = "https://your-naming-tool.azurewebsites.net"
  # apikey = "your-api-key"
}

# Fetch the azurerm resource type → short name mapping in one call.
# The `resources` attribute is a map(string):
#   key   = azurerm resource type  (e.g. "azurerm_storage_account")
#   value = short name             (e.g. "st")
data "proactnaming_azurerm_resources" "mapping" {}

# Look up the short name for a specific azurerm resource type.
locals {
  storage_short_name = data.proactnaming_azurerm_resources.mapping.resources["azurerm_storage_account"]
}

# Use the short name directly in a proactnaming_generate_name resource.
resource "proactnaming_generate_name" "storage" {
  organization  = "myorg"
  resource_type = local.storage_short_name
  application   = "webapp"
  function      = "data"
  instance      = "001"
  location      = "euw"
  environment   = "dev"
}

output "all_azurerm_mappings" {
  description = "Full map of azurerm resource type to short name."
  value       = data.proactnaming_azurerm_resources.mapping.resources
}

output "storage_account_short_name" {
  description = "Short name for azurerm_storage_account."
  value       = local.storage_short_name
}

output "storage_account_generated_name" {
  description = "Generated name for the storage account."
  value       = proactnaming_generate_name.storage.resource_name
}

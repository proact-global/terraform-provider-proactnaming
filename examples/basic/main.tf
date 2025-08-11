terraform {
  required_providers {
    proactnaming = {
      source  = "registry.terraform.io/proact-global/proactnaming"
      version = "~> 1.0"
    }
  }
}

# Configure the Proact Naming provider
# Update these values with your Azure Naming Tool details
provider "proactnaming" {
  host           = "https://mangopato-namingtool.azurewebsites.net/"
  apikey         = "6f510d50-ba62-40e1-a432-c3fc0a530483"
  admin_password = "Pa$$w0rd!"
}

# Generate a resource group name
resource "proactnaming_generate_name" "resource_group" {
  organization  = "man"
  resource_type = "rg"
  application   = "app"
  function      = "api"
  instance      = "002"
  location      = "euw"
  environment   = "dev"
}

# Output the generated name for use in other resources
output "resource_group_name" {
  description = "Generated Azure resource group name"
  value       = proactnaming_generate_name.resource_group.resource_name
}

# Example: Use the generated name in an Azure resource group
# resource "azurerm_resource_group" "main" {
#   name     = proactnaming_generate_name.resource_group.resource_name
#   location = "West Europe"
# }

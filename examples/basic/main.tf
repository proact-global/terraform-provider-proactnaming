terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

# Configure the Proact Naming provider
# Update these values with your Azure Naming Tool details
provider "proactnaming" {
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = "your-api-key"
  admin_password = "your-admin-password" # Optional, only needed for delete operations
}

# Generate a resource group name
resource "proactnaming_generate_name" "resource_group" {
  organization  = "myorg"
  resource_type = "rg"
  application   = "webapp"
  function      = "api"
  instance      = "001"
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

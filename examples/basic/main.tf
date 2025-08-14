# Copyright (c) HashiCorp, Inc.

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
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = "your-api-key-here"
  admin_password = "your-admin-password" # Only required for delete operations
}

# Generate a resource group name
resource "proactnaming_generate_name" "resource_group" {
  organization  = "myorg"  # Your organization code
  resource_type = "rg"     # Resource group
  application   = "webapp" # Application name
  function      = "api"    # Function/purpose
  instance      = "001"    # Instance number
  location      = "euw"    # Europe West
  environment   = "dev"    # Development environment
}

# Output the generated name for use in other resources
output "resource_group_name" {
  description = "Generated Azure resource group name"
  value       = proactnaming_generate_name.resource_group.resource_name
}

# Example: Use the generated name in an Azure resource group
# Uncomment and configure the AzureRM provider to use this
# resource "azurerm_resource_group" "main" {
#   name     = proactnaming_generate_name.resource_group.resource_name
#   location = "West Europe"
#   
#   tags = {
#     Environment = "dev"
#     Application = "webapp"
#   }
# }

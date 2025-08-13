# Copyright (c) HashiCorp, Inc.

# Mixed Configuration Example
# This approach allows overriding environment variables with Terraform variables

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

variable "naming_tool_host" {
  description = "Azure Naming Tool host (overrides PROACTNAMING_HOST env var)"
  type        = string
  default     = null
}

variable "naming_tool_apikey" {
  description = "Azure Naming Tool API key (overrides PROACTNAMING_APIKEY env var)"
  type        = string
  sensitive   = true
  default     = null
}

variable "naming_tool_admin_password" {
  description = "Azure Naming Tool admin password (overrides PROACTNAMING_ADMIN_PASSWORD env var)"
  type        = string
  sensitive   = true
  default     = null
}

provider "proactnaming" {
  # These values override environment variables if provided
  host           = var.naming_tool_host
  apikey         = var.naming_tool_apikey
  admin_password = var.naming_tool_admin_password
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

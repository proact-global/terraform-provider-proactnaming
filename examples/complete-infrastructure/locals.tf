# Copyright (c) HashiCorp, Inc.

locals {
  # Common tags to be applied to all resources
  common_tags = merge(var.project_tags, {
    Environment  = var.environment
    Application  = var.application
    Organization = var.organization
    Location     = var.location
    DeployedBy   = "Terraform"
    DeployedAt   = timestamp()
  })

  # Resource naming components
  naming_components = {
    organization = var.organization
    application  = var.application
    location     = var.location
    environment  = var.environment
  }

  # Network configuration
  network_config = {
    vnet_address_space = "10.0.0.0/16"
    web_subnet         = "10.0.1.0/24"
    app_subnet         = "10.0.2.0/24"
    data_subnet        = "10.0.3.0/24"
  }

  # Environment-specific configurations
  env_config = {
    dev = {
      vm_size             = "Standard_B1s"
      sql_sku             = "Basic"
      app_service_sku     = "B1"
      storage_replication = "LRS"
      zone_redundant      = false
    }
    test = {
      vm_size             = "Standard_B2s"
      sql_sku             = "S1"
      app_service_sku     = "S1"
      storage_replication = "LRS"
      zone_redundant      = false
    }
    stage = {
      vm_size             = "Standard_D2s_v3"
      sql_sku             = "S2"
      app_service_sku     = "P1v3"
      storage_replication = "GRS"
      zone_redundant      = false
    }
    prod = {
      vm_size             = "Standard_D4s_v3"
      sql_sku             = "S3"
      app_service_sku     = "P2v3"
      storage_replication = "GRS"
      zone_redundant      = true
    }
  }

  # Current environment configuration
  current_env = local.env_config[var.environment]
}

# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

# Configure the Proact Naming provider
provider "proactnaming" {
  host           = var.naming_tool_host
  apikey         = var.naming_tool_apikey
  admin_password = var.naming_tool_admin_password
}

# Infrastructure Resources
resource "proactnaming_generate_name" "resource_group" {
  organization  = var.organization
  resource_type = "rg"
  application   = var.application
  function      = "infra"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "proactnaming_generate_name" "storage_account" {
  organization  = var.organization
  resource_type = "st"
  application   = var.application
  function      = "data"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "proactnaming_generate_name" "virtual_network" {
  organization  = var.organization
  resource_type = "vnet"
  application   = var.application
  function      = "core"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

# Compute Resources
resource "proactnaming_generate_name" "kubernetes_cluster" {
  organization  = var.organization
  resource_type = "aks"
  application   = var.application
  function      = "web"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "proactnaming_generate_name" "function_app" {
  organization  = var.organization
  resource_type = "func"
  application   = var.application
  function      = "api"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

# Database Resources
resource "proactnaming_generate_name" "sql_server" {
  organization  = var.organization
  resource_type = "sql"
  application   = var.application
  function      = "db"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

# Monitoring Resources
resource "proactnaming_generate_name" "log_analytics" {
  organization  = var.organization
  resource_type = "log"
  application   = var.application
  function      = "mon"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

# Multi-instance examples
resource "proactnaming_generate_name" "key_vaults" {
  count = 3

  organization  = var.organization
  resource_type = "kv"
  application   = var.application
  function      = "sec"
  instance      = format("%03d", count.index + 1)
  location      = var.location
  environment   = var.environment
}

# Subnet with custom components (subnet_tier and subnet_instance)
resource "proactnaming_generate_name" "subnet_app" {
  organization  = var.organization
  resource_type = "snet"
  application   = var.application
  instance      = "001"
  location      = var.location
  environment   = var.environment

  custom_component {
    name  = "subnet_tier"
    value = "app"
  }

  custom_component {
    name  = "subnet_instance"
    value = "001"
  }
}

# Subnet using a dynamic block (useful when tiers are driven by a variable)
locals {
  subnet_tiers = [
    { name = "subnet_tier", value = "web" },
    { name = "subnet_instance", value = "001" },
  ]
}

resource "proactnaming_generate_name" "subnet_web" {
  organization  = var.organization
  resource_type = "snet"
  application   = var.application
  instance      = "001"
  location      = var.location
  environment   = var.environment

  dynamic "custom_component" {
    for_each = local.subnet_tiers
    content {
      name  = custom_component.value.name
      value = custom_component.value.value
    }
  }
}

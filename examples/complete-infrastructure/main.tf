# Copyright (c) HashiCorp, Inc.

terraform {
  required_version = ">= 1.0"

  required_providers {
    proactnaming = {
      source  = "proact-global/proactnaming"
      version = "~> 1.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

# Configure providers
provider "proactnaming" {
  # Configuration via environment variables
}

provider "azurerm" {
  features {}
}

# Get available resource types for validation
data "proactnaming_resource_types" "all" {}

# Validate required resource types are available
locals {
  enabled_types  = [for rt in data.proactnaming_resource_types.all.resource_types : rt.short_name if rt.enabled]
  required_types = ["rg", "st", "app", "plan", "sql", "sqldb", "vnet", "snet", "nsg", "pip", "vm"]

  missing_types = [for required in local.required_types : required if !contains(local.enabled_types, required)]

  # Fail if any required types are missing
  validation_check = length(local.missing_types) == 0 ? true : error("Missing required resource types: ${join(", ", local.missing_types)}")
}

# Infrastructure Resource Group
resource "proactnaming_generate_name" "infrastructure_rg" {
  organization  = var.organization
  resource_type = "rg"
  application   = var.application
  function      = "infra"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_resource_group" "infrastructure" {
  name     = proactnaming_generate_name.infrastructure_rg.resource_name
  location = var.azure_region

  tags = local.common_tags
}

# Application Resource Group
resource "proactnaming_generate_name" "application_rg" {
  organization  = var.organization
  resource_type = "rg"
  application   = var.application
  function      = "app"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_resource_group" "application" {
  name     = proactnaming_generate_name.application_rg.resource_name
  location = var.azure_region

  tags = local.common_tags
}

# Virtual Network
resource "proactnaming_generate_name" "vnet" {
  organization  = var.organization
  resource_type = "vnet"
  application   = var.application
  function      = "main"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_virtual_network" "main" {
  name                = proactnaming_generate_name.vnet.resource_name
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.infrastructure.location
  resource_group_name = azurerm_resource_group.infrastructure.name

  tags = local.common_tags
}

# Subnets
resource "proactnaming_generate_name" "subnet_web" {
  organization  = var.organization
  resource_type = "snet"
  application   = var.application
  function      = "web"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_subnet" "web" {
  name                 = proactnaming_generate_name.subnet_web.resource_name
  resource_group_name  = azurerm_resource_group.infrastructure.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "proactnaming_generate_name" "subnet_app" {
  organization  = var.organization
  resource_type = "snet"
  application   = var.application
  function      = "app"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_subnet" "app" {
  name                 = proactnaming_generate_name.subnet_app.resource_name
  resource_group_name  = azurerm_resource_group.infrastructure.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "proactnaming_generate_name" "subnet_data" {
  organization  = var.organization
  resource_type = "snet"
  application   = var.application
  function      = "data"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_subnet" "data" {
  name                 = proactnaming_generate_name.subnet_data.resource_name
  resource_group_name  = azurerm_resource_group.infrastructure.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.3.0/24"]
}

# Network Security Groups
resource "proactnaming_generate_name" "nsg_web" {
  organization  = var.organization
  resource_type = "nsg"
  application   = var.application
  function      = "web"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_network_security_group" "web" {
  name                = proactnaming_generate_name.nsg_web.resource_name
  location            = azurerm_resource_group.infrastructure.location
  resource_group_name = azurerm_resource_group.infrastructure.name

  security_rule {
    name                       = "HTTP"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "HTTPS"
    priority                   = 1002
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = local.common_tags
}

# Storage Account for application data
resource "proactnaming_generate_name" "storage" {
  organization  = var.organization
  resource_type = "st"
  application   = var.application
  function      = "data"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_storage_account" "data" {
  name                     = proactnaming_generate_name.storage.resource_name
  resource_group_name      = azurerm_resource_group.application.name
  location                 = azurerm_resource_group.application.location
  account_tier             = "Standard"
  account_replication_type = local.current_env.storage_replication

  tags = local.common_tags
} # App Service Plan
resource "proactnaming_generate_name" "app_service_plan" {
  organization  = var.organization
  resource_type = "plan"
  application   = var.application
  function      = "web"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_service_plan" "main" {
  name                = proactnaming_generate_name.app_service_plan.resource_name
  location            = azurerm_resource_group.application.location
  resource_group_name = azurerm_resource_group.application.name
  os_type             = "Linux"
  sku_name            = local.current_env.app_service_sku

  tags = local.common_tags
} # Web App
resource "proactnaming_generate_name" "web_app" {
  organization  = var.organization
  resource_type = "app"
  application   = var.application
  function      = "web"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_linux_web_app" "main" {
  name                = proactnaming_generate_name.web_app.resource_name
  location            = azurerm_resource_group.application.location
  resource_group_name = azurerm_resource_group.application.name
  service_plan_id     = azurerm_service_plan.main.id

  site_config {
    application_stack {
      node_version = "18-lts"
    }
  }

  app_settings = {
    "ENVIRONMENT" = var.environment
    "APPLICATION" = var.application
  }

  tags = local.common_tags
}

# SQL Server
resource "proactnaming_generate_name" "sql_server" {
  organization  = var.organization
  resource_type = "sql"
  application   = var.application
  function      = "main"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_mssql_server" "main" {
  name                         = proactnaming_generate_name.sql_server.resource_name
  resource_group_name          = azurerm_resource_group.application.name
  location                     = azurerm_resource_group.application.location
  version                      = "12.0"
  administrator_login          = var.sql_admin_username
  administrator_login_password = var.sql_admin_password

  tags = local.common_tags
}

# SQL Database
resource "proactnaming_generate_name" "sql_database" {
  organization  = var.organization
  resource_type = "sqldb"
  application   = var.application
  function      = "main"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

resource "azurerm_mssql_database" "main" {
  name           = proactnaming_generate_name.sql_database.resource_name
  server_id      = azurerm_mssql_server.main.id
  collation      = "SQL_Latin1_General_CP1_CI_AS"
  license_type   = "LicenseIncluded"
  sku_name       = local.current_env.sql_sku
  zone_redundant = local.current_env.zone_redundant

  tags = local.common_tags
} # Virtual Machines for additional workloads
resource "proactnaming_generate_name" "vm_worker" {
  count = var.worker_vm_count

  organization  = var.organization
  resource_type = "vm"
  application   = var.application
  function      = "worker"
  instance      = format("%03d", count.index + 1)
  location      = var.location
  environment   = var.environment
}

resource "azurerm_public_ip" "vm_worker" {
  count = var.worker_vm_count

  name                = "${proactnaming_generate_name.vm_worker[count.index].resource_name}-pip"
  location            = azurerm_resource_group.infrastructure.location
  resource_group_name = azurerm_resource_group.infrastructure.name
  allocation_method   = "Static"

  tags = local.common_tags
}

resource "azurerm_network_interface" "vm_worker" {
  count = var.worker_vm_count

  name                = "${proactnaming_generate_name.vm_worker[count.index].resource_name}-nic"
  location            = azurerm_resource_group.infrastructure.location
  resource_group_name = azurerm_resource_group.infrastructure.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.app.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.vm_worker[count.index].id
  }

  tags = local.common_tags
}

resource "azurerm_linux_virtual_machine" "worker" {
  count = var.worker_vm_count

  name                = proactnaming_generate_name.vm_worker[count.index].resource_name
  location            = azurerm_resource_group.infrastructure.location
  resource_group_name = azurerm_resource_group.infrastructure.name
  size                = local.current_env.vm_size
  admin_username      = "adminuser"

  disable_password_authentication = true

  network_interface_ids = [
    azurerm_network_interface.vm_worker[count.index].id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = var.ssh_public_key
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  tags = local.common_tags
}

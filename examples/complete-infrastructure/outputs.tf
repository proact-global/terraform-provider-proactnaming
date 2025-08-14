# Copyright (c) HashiCorp, Inc.

# Resource Names
output "resource_names" {
  description = "Map of all generated resource names"
  value = {
    infrastructure_rg = proactnaming_generate_name.infrastructure_rg.resource_name
    application_rg    = proactnaming_generate_name.application_rg.resource_name
    vnet              = proactnaming_generate_name.vnet.resource_name
    subnet_web        = proactnaming_generate_name.subnet_web.resource_name
    subnet_app        = proactnaming_generate_name.subnet_app.resource_name
    subnet_data       = proactnaming_generate_name.subnet_data.resource_name
    nsg_web           = proactnaming_generate_name.nsg_web.resource_name
    storage_account   = proactnaming_generate_name.storage.resource_name
    app_service_plan  = proactnaming_generate_name.app_service_plan.resource_name
    web_app           = proactnaming_generate_name.web_app.resource_name
    sql_server        = proactnaming_generate_name.sql_server.resource_name
    sql_database      = proactnaming_generate_name.sql_database.resource_name
    worker_vms        = [for vm in proactnaming_generate_name.vm_worker : vm.resource_name]
  }
}

# Resource Group Information
output "resource_group_ids" {
  description = "Azure resource group IDs"
  value = {
    infrastructure = azurerm_resource_group.infrastructure.id
    application    = azurerm_resource_group.application.id
  }
}

# Network Information
output "network_info" {
  description = "Network configuration details"
  value = {
    vnet_id            = azurerm_virtual_network.main.id
    vnet_address_space = azurerm_virtual_network.main.address_space
    subnet_ids = {
      web  = azurerm_subnet.web.id
      app  = azurerm_subnet.app.id
      data = azurerm_subnet.data.id
    }
  }
}

# Application URLs
output "application_urls" {
  description = "URLs for deployed applications"
  value = {
    web_app = "https://${azurerm_linux_web_app.main.default_hostname}"
  }
}

# Database Connection Information
output "database_info" {
  description = "Database connection information"
  value = {
    sql_server_fqdn   = azurerm_mssql_server.main.fully_qualified_domain_name
    sql_database_name = azurerm_mssql_database.main.name
  }
  sensitive = false
}

# Storage Account Information
output "storage_info" {
  description = "Storage account information"
  value = {
    storage_account_name        = azurerm_storage_account.data.name
    storage_account_primary_url = azurerm_storage_account.data.primary_blob_endpoint
  }
}

# Virtual Machine Information
output "vm_info" {
  description = "Virtual machine information"
  value = {
    worker_vm_names = [for vm in azurerm_linux_virtual_machine.worker : vm.name]
    worker_vm_ips   = [for pip in azurerm_public_ip.vm_worker : pip.ip_address]
  }
}

# Environment Configuration
output "environment_info" {
  description = "Environment-specific configuration"
  value = {
    environment     = var.environment
    location        = var.location
    azure_region    = var.azure_region
    worker_vm_count = var.worker_vm_count
    current_config  = local.current_env
  }
}

# Cost Management
output "cost_tags" {
  description = "Tags for cost management and tracking"
  value = {
    common_tags = local.common_tags
    cost_center = var.project_tags["CostCenter"]
  }
}

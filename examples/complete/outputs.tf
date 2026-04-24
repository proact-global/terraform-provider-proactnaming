# Copyright (c) HashiCorp, Inc.

# Infrastructure Resource Names
output "infrastructure_names" {
  description = "Generated names for infrastructure resources"
  value = {
    resource_group  = proactnaming_generate_name.resource_group.resource_name
    storage_account = proactnaming_generate_name.storage_account.resource_name
    virtual_network = proactnaming_generate_name.virtual_network.resource_name
  }
}

# Compute Resource Names
output "compute_names" {
  description = "Generated names for compute resources"
  value = {
    kubernetes_cluster = proactnaming_generate_name.kubernetes_cluster.resource_name
    function_app       = proactnaming_generate_name.function_app.resource_name
  }
}

# Database Resource Names
output "database_names" {
  description = "Generated names for database resources"
  value = {
    sql_server = proactnaming_generate_name.sql_server.resource_name
  }
}

# Monitoring Resource Names
output "monitoring_names" {
  description = "Generated names for monitoring resources"
  value = {
    log_analytics = proactnaming_generate_name.log_analytics.resource_name
  }
}

# Multi-instance Resource Names
output "key_vault_names" {
  description = "Generated names for key vaults"
  value       = [for kv in proactnaming_generate_name.key_vaults : kv.resource_name]
}

# All Generated Names (for easy reference)
output "all_resource_names" {
  description = "All generated resource names organized by category"
  value = {
    infrastructure = {
      resource_group  = proactnaming_generate_name.resource_group.resource_name
      storage_account = proactnaming_generate_name.storage_account.resource_name
      virtual_network = proactnaming_generate_name.virtual_network.resource_name
    }
    compute = {
      kubernetes_cluster = proactnaming_generate_name.kubernetes_cluster.resource_name
      function_app       = proactnaming_generate_name.function_app.resource_name
      key_vaults         = [for kv in proactnaming_generate_name.key_vaults : kv.resource_name]
    }
    database = {
      sql_server = proactnaming_generate_name.sql_server.resource_name
    }
    monitoring = {
      log_analytics = proactnaming_generate_name.log_analytics.resource_name
    }
  }
}

# Copyright (c) HashiCorp, Inc.

# Data source outputs
output "all_resource_types" {
  description = "List of all available resource types"
  value       = data.proactnaming_resourcetypes.all.resource_types
}

# Data source outputs
output "all_resource_types" {
  description = "List of all available resource types"
  value       = data.proactnaming_resourcetypes.all.resource_types
}

output "available_types_summary" {
  description = "Summary of available resource types"
  value = {
    total_count     = length(data.proactnaming_resourcetypes.all.resource_types)
    available_types = local.available_types
    missing_types   = local.missing_types
  }
}

# Generated name outputs (when using the conditional resource)
output "conditional_resource_name" {
  description = "The generated resource group name (if created)"
  value       = length(proactnaming_generate_name.conditional_rg) > 0 ? proactnaming_generate_name.conditional_rg[0].generated_name[0].resource_name : "Not created - missing required resource types"
}

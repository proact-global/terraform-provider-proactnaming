# A configuration for trying the provider locally against a naming tool,
# without publishing a release or installing anything from the registry.
#
# Run it with scripts/local-dev.sh, which builds the provider and points
# Terraform at the binary through a development override.

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

# Credentials come from PROACTNAMING_HOST, PROACTNAMING_APIKEY and
# PROACTNAMING_ADMIN_PASSWORD, so none appear here.
provider "proactnaming" {
  # Set to true to work the name out from the naming tool's configuration
  # instead of generating and deleting a preview entry while planning.
  # Planning then writes nothing at all and needs no admin password.
  predict_names_locally = var.predict_names_locally
}

variable "predict_names_locally" {
  description = "Work the plan-time name out locally rather than generating a preview entry."
  type        = bool
  default     = false
}

variable "organization" {
  type    = string
  default = "pca"
}

variable "location" {
  type    = string
  default = "we"
}

variable "environment" {
  type    = string
  default = "p"
}

variable "resource_type" {
  description = "A resource type short name this naming tool defines."
  type        = string
  default     = "st"
}

resource "proactnaming_generate_name" "example" {
  organization  = var.organization
  resource_type = var.resource_type
  application   = "localdev"
  instance      = "001"
  location      = var.location
  environment   = var.environment
}

# What the naming tool is configured to accept, which is also what the
# provider checks a request against while planning.
data "proactnaming_resource_types" "all" {}

# azurerm resource type to short name, so a name can be requested without
# hard-coding the abbreviation.
data "proactnaming_azurerm_resources" "lookup" {}

output "generated_name" {
  description = "The name the naming tool produced."
  value       = proactnaming_generate_name.example.resource_name
}

output "record_id" {
  description = "Its id in the naming tool, which the provider uses to detect deletion."
  value       = proactnaming_generate_name.example.id
}

output "resource_type_count" {
  value = length(data.proactnaming_resource_types.all.resource_types)
}

output "storage_account_short_name" {
  value = try(data.proactnaming_azurerm_resources.lookup.resources["azurerm_storage_account"], "not configured")
}

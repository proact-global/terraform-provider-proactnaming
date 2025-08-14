# Copyright (c) HashiCorp, Inc.

variable "organization" {
  description = "Organization identifier for naming"
  type        = string
  default     = "myorg"

  validation {
    condition     = length(var.organization) >= 2 && length(var.organization) <= 10
    error_message = "Organization must be between 2 and 10 characters."
  }
}

variable "application" {
  description = "Application identifier for naming"
  type        = string
  default     = "webapp"

  validation {
    condition     = length(var.application) >= 2 && length(var.application) <= 15
    error_message = "Application must be between 2 and 15 characters."
  }
}

variable "location" {
  description = "Azure region identifier for naming (e.g., 'euw', 'eus')"
  type        = string
  default     = "euw"

  validation {
    condition     = contains(["euw", "eus", "weu", "neu", "sea", "eas"], var.location)
    error_message = "Location must be a valid Azure region identifier."
  }
}

variable "azure_region" {
  description = "Full Azure region name for resource deployment"
  type        = string
  default     = "West Europe"

  validation {
    condition = contains([
      "West Europe", "North Europe", "East US", "West US", "Central US",
      "South Central US", "East US 2", "West US 2", "Southeast Asia", "East Asia"
    ], var.azure_region)
    error_message = "Azure region must be a valid region name."
  }
}

variable "environment" {
  description = "Environment identifier (dev, test, stage, prod)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "test", "stage", "prod"], var.environment)
    error_message = "Environment must be one of: dev, test, stage, prod."
  }
}

variable "worker_vm_count" {
  description = "Number of worker VMs to create"
  type        = number
  default     = 2

  validation {
    condition     = var.worker_vm_count >= 0 && var.worker_vm_count <= 10
    error_message = "Worker VM count must be between 0 and 10."
  }
}

variable "worker_vm_size" {
  description = "Size of worker VMs"
  type        = string
  default     = "Standard_B2s"

  validation {
    condition = contains([
      "Standard_B1s", "Standard_B1ms", "Standard_B2s", "Standard_B2ms",
      "Standard_D2s_v3", "Standard_D4s_v3", "Standard_F2s_v2", "Standard_F4s_v2"
    ], var.worker_vm_size)
    error_message = "Worker VM size must be a valid Azure VM size."
  }
}

variable "sql_admin_username" {
  description = "SQL Server administrator username"
  type        = string
  default     = "sqladmin"
  sensitive   = true

  validation {
    condition     = length(var.sql_admin_username) >= 4 && length(var.sql_admin_username) <= 20
    error_message = "SQL admin username must be between 4 and 20 characters."
  }
}

variable "sql_admin_password" {
  description = "SQL Server administrator password"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.sql_admin_password) >= 8
    error_message = "SQL admin password must be at least 8 characters long."
  }
}

variable "ssh_public_key" {
  description = "SSH public key for VM access"
  type        = string

  validation {
    condition     = can(regex("^ssh-", var.ssh_public_key))
    error_message = "SSH public key must start with 'ssh-'."
  }
}

variable "project_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default = {
    "Project"    = "Infrastructure Demo"
    "ManagedBy"  = "Terraform"
    "CostCenter" = "IT"
  }
}

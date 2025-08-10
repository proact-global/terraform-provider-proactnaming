variable "naming_tool_host" {
  description = "The base URL for the Azure Naming Tool API"
  type        = string
  default     = "https://your-naming-tool.azurewebsites.net"
}

variable "naming_tool_apikey" {
  description = "API key for authenticating with the Azure Naming Tool"
  type        = string
  sensitive   = true
}

variable "naming_tool_admin_password" {
  description = "Admin password for delete operations (optional)"
  type        = string
  sensitive   = true
  default     = null
}

variable "organization" {
  description = "Organization identifier"
  type        = string
  default     = "myorg"
}

variable "application" {
  description = "Application identifier"
  type        = string
  default     = "webapp"
}

variable "location" {
  description = "Azure region identifier"
  type        = string
  default     = "euw"
}

variable "environment" {
  description = "Environment identifier"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "test", "stage", "prod"], var.environment)
    error_message = "Environment must be one of: dev, test, stage, prod."
  }
}

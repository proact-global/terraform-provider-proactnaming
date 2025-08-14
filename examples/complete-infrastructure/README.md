# Comprehensive Azure Infrastructure Example

This example demonstrates a complete Azure infrastructure deployment using the proactnaming provider to generate standardized names for all resources.

## Prerequisites

- Azure subscription and AzureRM provider configured
- Access to an Azure Naming Tool instance
- Valid API key for the naming tool

## Features Demonstrated

- **Multi-tier architecture**: Web, app, and database tiers
- **Multiple environments**: Production and staging
- **Resource scaling**: Multiple instances with proper naming
- **Data source usage**: Validation and resource type discovery
- **Best practices**: Environment variables and validation

## Configuration

Set these environment variables before running:

```bash
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="your-api-key"
export PROACTNAMING_ADMIN_PASSWORD="your-admin-password"  # Optional
export ARM_SUBSCRIPTION_ID="your-azure-subscription-id"
export ARM_CLIENT_ID="your-azure-client-id"
export ARM_CLIENT_SECRET="your-azure-client-secret"
export ARM_TENANT_ID="your-azure-tenant-id"
```

## Usage

```bash
terraform init
terraform plan -var="environment=prod"
terraform apply -var="environment=prod"
```

## Outputs

- `resource_names` - Map of all generated resource names
- `resource_group_id` - Azure resource group ID
- `application_urls` - URLs for deployed applications

# Proact Naming Provider Examples

This directory contains examples demonstrating how to use the Proact Naming Terraform provider to generate standardized Azure resource names using the Azure Naming Tool.

## Examples Overview

### 🚀 [basic/](./basic/)
Simple example showing basic resource name generation with minimal configuration.
- Single resource group name generation
- Basic provider configuration
- Output usage demonstration

### 🏢 [complete/](./complete/)
Comprehensive example demonstrating all features including:
- Multiple resource types
- Different environments
- Data sources usage
- Output formatting

### 🏗️ [complete-infrastructure/](./complete-infrastructure/)
Full Azure infrastructure deployment example with:
- Multi-tier architecture (web, app, database)
- Multiple environments with different configurations
- Virtual machines with scaling
- Networking components (VNet, subnets, NSGs)
- Storage accounts and SQL databases
- Environment-specific resource sizing
- Comprehensive tagging strategy

### 🔧 [provider-configuration/](./provider-configuration/)
Examples of different provider configuration methods:
- Using environment variables (recommended)
- Direct configuration (for testing)
- Mixed approaches

### 📊 [data-sources/](./data-sources/)
Examples using the provider's data sources:
- Querying available resource types
- Filtering and validation logic
- Looking up existing generated names
- Dynamic resource creation based on availability

## Prerequisites

1. **Azure Naming Tool**: You need access to an Azure Naming Tool instance
2. **API Key**: Valid API key for the Azure Naming Tool
3. **Admin Password**: (Optional) Required only for delete operations

## Quick Start

1. Copy one of the example directories
2. Update the provider configuration with your Azure Naming Tool details:
   ```hcl
   provider "proactnaming" {
     host     = "https://your-naming-tool.azurewebsites.net"
     apikey   = "your-api-key"
     admin_password = "your-admin-password"  # Optional, for deletes
   }
   ```
3. Run `terraform init` and `terraform plan`

## Environment Variables

You can also configure the provider using environment variables:

```bash
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="your-api-key"
export PROACTNAMING_ADMIN_PASSWORD="your-admin-password"  # Optional
```

## Common Use Cases

- **Infrastructure Naming**: Generate consistent names for Azure resources
- **Multi-Environment**: Maintain naming consistency across dev/test/prod
- **Team Collaboration**: Ensure all team members follow naming standards
- **Compliance**: Meet organizational naming requirements automatically

## Support

For issues and questions:
- Check the [provider documentation](../docs/)
- Review the [Azure Naming Tool documentation](https://github.com/microsoft/CloudAdoptionFramework/tree/master/ready/AzNamingTool)

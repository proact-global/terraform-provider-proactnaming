# Data Sources Example

This example demonstrates how to use the proactnaming provider's data sources to:

1. **Retrieve available resource types** - Get a list of all supported Azure resource types from your naming tool
2. **Query existing generated names** - Retrieve details of previously generated names by ID
3. **Validate configurations** - Use data sources to validate resource types before generating names

## Prerequisites

- Access to an Azure Naming Tool instance
- Valid API key for the naming tool
- (Optional) Admin password for delete operations

## Configuration

1. Update the provider configuration with your naming tool details:
   ```hcl
   provider "proactnaming" {
     host           = "https://your-naming-tool.azurewebsites.net"
     apikey         = "your-api-key-here"
     admin_password = "your-admin-password" # Optional
   }
   ```

2. If querying an existing generated name, update the ID:
   ```hcl
   data "proactnaming_generated_name" "existing" {
     id = 123 # Replace with actual ID
   }
   ```

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Outputs

- `available_resource_types` - Complete list of supported resource types
- `existing_name_details` - Details of the queried generated name

## Features Demonstrated

- **Resource Type Discovery**: Query all available resource types
- **Name Retrieval**: Get details of existing generated names
- **Validation Logic**: Use data sources to validate configurations
- **Conditional Resources**: Create resources based on validation results

## Available Data Sources

### 1. Resource Types (`proactnaming_resource_types`)
Query all available resource types configured in your Azure Naming Tool.

### 2. Generated Names (`proactnaming_generated_name`)
Query specific generated names by their ID.

- `resource-types.tf` - Examples using the resource types data source
- `generated-names.tf` - Examples using the generated names data source
- `outputs.tf` - Structured outputs for easy consumption

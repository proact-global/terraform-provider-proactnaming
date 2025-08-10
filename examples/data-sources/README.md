# Data Sources Examples

This directory contains examples of using the Proact Naming provider's data sources to query information from the Azure Naming Tool.

## Available Data Sources

### 1. Resource Types (`proactnaming_resourcetypes`)
Query all available resource types configured in your Azure Naming Tool.

### 2. Generated Names (`proactnaming_generated_name`)
Query specific generated names by their ID.

## Use Cases

- **Discovery**: Find out what resource types are available
- **Validation**: Check if certain resource types exist before using them
- **Reference**: Look up previously generated names
- **Debugging**: Investigate naming tool configuration

## Files

- `resource-types.tf` - Examples using the resource types data source
- `generated-names.tf` - Examples using the generated names data source
- `outputs.tf` - Structured outputs for easy consumption

# Complete Example

This comprehensive example demonstrates all features of the Proact Naming provider including:

- Multiple resource types (Resource Group, Storage Account, Virtual Machine, etc.)
- Different environments (dev, test, prod)
- Using data sources to query resource types
- Organized outputs for different resource categories
- Best practices for resource naming

## What This Example Creates

- **Infrastructure Names**: Resource groups, storage accounts, networking
- **Compute Names**: Virtual machines, app services
- **Data Names**: Databases, analytics resources
- **Multi-Environment**: Same pattern across dev/test/prod

## Features Demonstrated

1. **Multiple Resource Types**: Shows naming for various Azure resources
2. **Environment Variations**: Same naming pattern across environments
3. **Data Sources**: Uses data sources to query available resource types
4. **Output Organization**: Structured outputs for easy consumption
5. **Real-World Patterns**: Practical naming scenarios

## Usage

1. Update provider configuration with your Azure Naming Tool details
2. Adjust the variables in `variables.tf` for your organization
3. Run:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## File Structure

- `main.tf` - Main resource declarations
- `variables.tf` - Input variables
- `outputs.tf` - Organized outputs
- `data.tf` - Data source queries
- `README.md` - This documentation

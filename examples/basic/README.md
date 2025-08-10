# Basic Example

This is a simple example demonstrating basic usage of the Proact Naming provider to generate an Azure resource group name.

## What This Example Does

- Configures the Proact Naming provider
- Generates a single resource group name
- Outputs the generated name for use in other resources

## Configuration

Update the provider configuration with your Azure Naming Tool details:

```hcl
provider "proactnaming" {
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = "your-api-key"
  admin_password = "your-admin-password"  # Optional, only for delete operations
}
```

## Usage

1. Clone this example
2. Update the provider configuration
3. Run:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Expected Output

The generated name will follow your Azure Naming Tool's configured pattern, for example:
- Input: `org=myorg, type=rg, app=webapp, instance=001, location=euw, env=dev`
- Output: `myorg-rg-webapp001-euw-dev`

## Notes

- All input fields trigger resource replacement when changed
- The generated name appears in `terraform plan` output (not "known after apply")
- No actual Azure resources are created - this only generates names

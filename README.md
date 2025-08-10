# Terraform Provider for ProAct Azure Naming Tool

This Terraform provider integrates with the Azure Naming Tool to generate standardized Azure resource names following organizational naming conventions.

## Features

- **🎯 Plan Visibility**: Shows actual generated names during `terraform plan` instead of "(known after apply)"
- **🧹 Clean Resource Management**: Automatic cleanup prevents accumulation in the Azure Naming Tool
- **🔄 Smart Replacements**: Changes to naming inputs trigger proper resource replacement
- **🛡️ Secure Configuration**: Sensitive credentials properly handled
- **📋 Complete Lifecycle**: Full CRUD operations with proper state management

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23 (for development)
- Access to an Azure Naming Tool instance
- Valid API credentials for the Azure Naming Tool

## Quick Start

### 1. Configure the Provider

```hcl
terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
      version = "~> 1.0"
    }
  }
}

provider "proactnaming" {
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = var.naming_tool_apikey
  admin_password = var.naming_tool_admin_password # Required for delete operations
}
```

### 2. Generate Resource Names

```hcl
resource "proactnaming_generate_name" "storage" {
  organization  = "myorg"
  resource_type = "st"
  application   = "webapp"
  function      = "data"
  instance      = "001"
  location      = "euw"
  environment   = "prod"
}

# Use the generated name in other resources
resource "azurerm_storage_account" "example" {
  name                = proactnaming_generate_name.storage.resource_name
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  # ... other configuration
}
```

### 3. Environment Variables (Alternative Configuration)

```bash
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="your-api-key"
export PROACTNAMING_ADMIN_PASSWORD="your-admin-password"
```

## Provider Configuration

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `host` | string | Yes | Base URL of your Azure Naming Tool instance |
| `apikey` | string | Yes | API key for authentication |
| `admin_password` | string | No | Admin password for delete operations |

## Resource: `proactnaming_generate_name`

### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `organization` | string | Yes | Organization identifier |
| `resource_type` | string | Yes | Azure resource type (e.g., 'rg', 'st', 'vm') |
| `application` | string | Yes | Application identifier |
| `function` | string | No | Function or purpose identifier |
| `instance` | string | Yes | Instance number or identifier |
| `location` | string | Yes | Azure region identifier |
| `environment` | string | Yes | Environment identifier |

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | number | Unique identifier in Azure Naming Tool |
| `resource_name` | string | Generated Azure resource name |
| `success` | boolean | Whether generation was successful |
| `message` | string | Message from the Azure Naming Tool |

## Examples

See the [examples](./examples/) directory for complete usage examples.

## Development

### Building the Provider

```shell
git clone https://github.com/proact-global/terraform-provider-proactnaming
cd terraform-provider-proactnaming
go build
```

### Running Tests

```shell
# Unit tests
go test ./...

# Acceptance tests (requires live Azure Naming Tool)
TF_ACC=1 go test ./internal/provider -v
```

## Support

For issues related to:
- **Provider functionality**: Open an issue in this repository
- **Azure Naming Tool**: Consult your Azure Naming Tool documentation
- **Terraform core**: Visit [Terraform's documentation](https://www.terraform.io/docs)

## License

This provider is released under the [MIT License](LICENSE).

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

# Using the Proact Naming Terraform Provider

The Proact Naming Terraform provider is deployed to our private Terrakube registry at `registry.eng.proactmcs.eu`. Due to GPG signature verification requirements, the recommended approach is to use a filesystem mirror setup.

## Quick Setup

### Automated Setup (Recommended)
Run the setup script from the provider repository:

```bash
./scripts/setup-local-provider.sh
```

This script will:
- Detect your platform (OS/architecture)
- Download the appropriate provider binary from Azure Storage
- Set up the filesystem mirror configuration
- Configure your `~/.terraformrc` file

### Manual Setup

1. **Download the provider binary** for your platform:
   - **macOS ARM64**: [terraform-provider-proactnaming_1.0.0_darwin_arm64.zip](https://proactdevelopment1231754.blob.core.windows.net/terraform-providers/proact/proactnaming/1.0.0/terraform-provider-proactnaming_1.0.0_darwin_arm64.zip)
   - **Linux AMD64**: [terraform-provider-proactnaming_1.0.0_linux_amd64.zip](https://proactdevelopment1231754.blob.core.windows.net/terraform-providers/proact/proactnaming/1.0.0/terraform-provider-proactnaming_1.0.0_linux_amd64.zip)

2. **Create the plugin directory**:
   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.eng.proactmcs.eu/proact/proactnaming/1.0.0/darwin_arm64
   # OR for Linux:
   # mkdir -p ~/.terraform.d/plugins/registry.eng.proactmcs.eu/proact/proactnaming/1.0.0/linux_amd64
   ```

3. **Extract and install the binary**:
   ```bash
   cd ~/.terraform.d/plugins/registry.eng.proactmcs.eu/proact/proactnaming/1.0.0/darwin_arm64
   unzip ~/Downloads/terraform-provider-proactnaming_1.0.0_darwin_arm64.zip
   chmod +x terraform-provider-proactnaming_v1.0.0
   ```

4. **Configure Terraform CLI** (`~/.terraformrc`):
   ```hcl
   provider_installation {
     filesystem_mirror {
       path = "~/.terraform.d/plugins"
       include = ["registry.eng.proactmcs.eu/proact/*"]
     }
     
     network_mirror {
       url = "https://registry.terraform.io/"
       include = ["registry.terraform.io/*/*"]
     }
   }
   ```

## Usage in Terraform

Add the provider to your `terraform` block:

```hcl
terraform {
  required_providers {
    proactnaming = {
      source  = "registry.eng.proactmcs.eu/proact/proactnaming"
      version = "~> 1.0"
    }
  }
}

provider "proactnaming" {
  host    = "https://your-naming-tool.azurewebsites.net"
  api_key = var.naming_tool_api_key
}
```

Then use the provider resources:

```hcl
resource "proactnaming_generate_name" "example" {
  resource_type = "rg"
  organization  = "myorg"
  location      = "euw"
  environment   = "dev"
  application   = "webapp"
  function      = "api"
  instance      = "001"
}

output "generated_name" {
  value = proactnaming_generate_name.example.resource_name
}
```

## Why Filesystem Mirror?

The filesystem mirror approach is used because:
- **Bypasses GPG signature verification** issues with private registries
- **Ensures consistent provider versions** across team members
- **Works offline** once the provider is cached locally
- **Faster initialization** as no network requests are needed for the provider

## Alternative: Development Overrides

For development work, you can also use development overrides:

```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "registry.eng.proactmcs.eu/proact/proactnaming" = "/path/to/your/built/provider"
  }
  
  direct {}
}
```

## Troubleshooting

### Provider Not Found
- Ensure the filesystem mirror path in `~/.terraformrc` is correct
- Verify the provider binary exists in the expected location
- Check that the binary is executable (`chmod +x`)

### Platform Mismatch
- Ensure you downloaded the correct binary for your platform
- The directory structure must match: `{os}_{arch}` (e.g., `darwin_arm64`, `linux_amd64`)

### Lock File Warnings
When using filesystem mirrors, you may see warnings about "unauthenticated" providers and incomplete lock files. This is expected and safe.

To generate lock file entries for additional platforms:
```bash
terraform providers lock -platform=linux_amd64 -platform=darwin_arm64
```

## Registry Information

- **Registry URL**: `registry.eng.proactmcs.eu`
- **Provider Source**: `registry.eng.proactmcs.eu/proact/proactnaming`
- **Current Version**: `1.0.0`
- **Supported Platforms**: 
  - `darwin/arm64` (macOS Apple Silicon)
  - `linux/amd64` (Linux x86_64)

## Support

For issues with the provider:
1. Check the [provider documentation](README.md)
2. Verify your Azure Naming Tool configuration
3. Contact the development team

---
page_title: "Using the Proact Naming provider"
subcategory: ""
description: |-
  Installing and configuring the provider, generating names, and resolving the
  errors the Azure Naming Tool returns.
---

# Using the Proact Naming Terraform Provider

The provider is published on the public [Terraform Registry](https://registry.terraform.io/providers/proact-global/proactnaming), so `terraform init` installs it for you. No manual download, filesystem mirror or plugin directory is needed.

## Installation

```hcl
terraform {
  required_providers {
    proactnaming = {
      source  = "proact-global/proactnaming"
      version = "~> 0.6"
    }
  }
}
```

Then run `terraform init`. Terraform verifies the release signature automatically.

## Configuring the provider

```hcl
provider "proactnaming" {
  host           = "https://your-naming-tool.azurewebsites.net"
  apikey         = var.naming_tool_apikey
  admin_password = var.naming_tool_admin_password
}
```

| Argument | Environment variable | Notes |
|----------|---------------------|-------|
| `host` | `PROACTNAMING_HOST` | Base URL of your Azure Naming Tool. Must include the scheme and must not end in a slash. |
| `apikey` | `PROACTNAMING_APIKEY` | API key with permission to generate names. |
| `admin_password` | `PROACTNAMING_ADMIN_PASSWORD` | Global admin password. See the note below — it is needed more often than you might expect. |
| `predict_names_locally` | `PROACTNAMING_PREDICT_NAMES_LOCALLY` | Optional, off by default. Work the plan-time name out locally rather than generating one. See below. |

All three can be supplied by environment variable instead, which keeps credentials out of your configuration:

```bash
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="…"
export PROACTNAMING_ADMIN_PASSWORD="…"
```

### The admin password is needed during `plan`, not just `destroy`

To show the real name during `terraform plan` rather than `(known after apply)`, the provider generates the name at plan time and immediately deletes that preview entry. Deleting is an Admin API call, so **`admin_password` must be available whenever anyone runs `terraform plan`** — including plan-only CI pipelines that would otherwise need no write credentials.

If the password is wrong, plans still appear to succeed: the naming tool answers an incorrect password with HTTP 200 and a `FAILURE` body rather than an authentication error. Since v0.5.0 the provider reports this as a warning naming the entry it could not remove. Treat that warning as "check `admin_password`".

## Generating a name

```hcl
resource "proactnaming_generate_name" "storage" {
  organization  = "myorg"
  resource_type = "st"
  application   = "webapp"
  function      = "data"   # optional
  instance      = "001"
  location      = "euw"
  environment   = "prod"
}

resource "azurerm_storage_account" "example" {
  name                = proactnaming_generate_name.storage.resource_name
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  # …
}
```

`organization`, `resource_type`, `application`, `instance`, `location` and `environment` are required; `function` is optional. Every input forces replacement when changed, so a generated name never silently drifts from the inputs that produced it.

Attributes available afterwards: `resource_name`, `id` (the record's ID in the naming tool), `success` and `message`.

The values you supply must exist in your naming tool's own configuration. `resource_type` is matched **exactly and case-sensitively** against the configured short names, and components such as `instance` may have a fixed length.

### Resource types needing extra components

Some resource types require components beyond the fixed arguments. Supply them as `custom_component` blocks:

```hcl
resource "proactnaming_generate_name" "subnet" {
  organization  = "myorg"
  resource_type = "snet"
  application   = "webapp"
  instance      = "001"
  location      = "euw"
  environment   = "prod"

  custom_component {
    name  = "subnet_tier"
    value = "app"
  }

  custom_component {
    name  = "subnet_instance"
    value = "01"
  }
}
```

Use a `dynamic "custom_component"` block when the set varies. Note that `application` is itself sent as a custom component, so a block named `application` overrides the `application` argument.

## Data sources

```hcl
# Every resource type your naming tool knows about, with its rules.
data "proactnaming_resource_types" "all" {}

# Map of azurerm resource type to short name, e.g. azurerm_storage_account => st.
data "proactnaming_azurerm_resources" "lookup" {}

resource "proactnaming_generate_name" "from_lookup" {
  resource_type = data.proactnaming_azurerm_resources.lookup.resources["azurerm_storage_account"]
  # …
}

# Look up a previously generated name by its ID.
data "proactnaming_generated_name" "existing" {
  id = proactnaming_generate_name.storage.id
}
```

`proactnaming_azurerm_resources` is the convenient one: it saves hard-coding short names, and it stays correct if your naming tool's configuration changes.

## Local development

To test a locally built provider, use `dev_overrides` rather than installing into a plugin directory:

```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "proact-global/proactnaming" = "/path/to/your/gobin"
  }

  direct {}
}
```

Build with `go install`, and Terraform will use your binary. `terraform init` is skipped for overridden providers — Terraform prints a warning to that effect, which is expected.

## Working the name out instead of generating one

To show a real name during `terraform plan` rather than `(known after apply)`, the provider generates the name at plan time and deletes the preview entry again. That is accurate by construction, but it writes to the naming tool on every plan, needs `admin_password` to clean up, and leaves a record behind whenever the cleanup fails.

Setting `predict_names_locally` avoids all of it. The provider reads the naming tool's configuration and works the name out from it, so planning performs no writes and needs no admin password:

```hcl
provider "proactnaming" {
  predict_names_locally = true
}
```

The saving is not marginal. Creating two resources consumed **ten** records in the naming tool with the default behaviour, eight of them previews created and deleted again, against **two** with this set.

The apply still calls the API, so the stored name is always the tool's own. If a worked-out name ever disagrees with it, the apply fails and reports both names rather than storing one the plan did not show.

### Limitations

**It does not work for every resource type.** Deciding the name requires evaluating the resource type's validation pattern, because that pattern is what decides whether the delimiter survives. Go's regular expressions cannot evaluate every pattern .NET accepts — notably lookarounds, written `(?!...)` or `(?=...)`. Azure's rule that a resource group name may not end with a period is expressed that way:

```
^(?!.*[\.]$)[a-zA-Z0-9_\.()-]{1,90}$
```

For such a type the provider reports that the name cannot be worked out and names the pattern, rather than guessing at one that might be contradicted at apply. Resource types with plainer patterns, such as a storage account's `^[a-z0-9]{3,24}$`, are unaffected.

So treat the setting as usable per resource type. If a plan fails saying a pattern cannot be evaluated, unset it, or keep those resources in a separate configuration.

**Re-check it after upgrading the Azure Naming Tool.** Working the name out reproduces the tool's algorithm. The reproduction reads configuration at run time, so it follows a deployment's own components, ordering and delimiter — but the algorithm itself is a copy, and a tool upgrade could change it. Upstream changes it roughly once a year, on major versions.

After an upgrade, run a plan and apply with the setting on and confirm they agree. A disagreement fails the apply with both names reported, so it cannot pass silently.

## Troubleshooting

**`host … is missing a scheme`**
`host` needs the protocol: `https://your-naming-tool.azurewebsites.net`, not `your-naming-tool.azurewebsites.net`. Earlier versions accepted this and failed later with an unhelpful `unsupported protocol scheme ""`.

**`Api Key is not valid!` / `Api Key was not provided!`**
`apikey` is wrong or unset.

**`Preview Name Cleanup Failed` warning**
The plan-time preview entry could not be removed. Almost always an incorrect `admin_password`. The plan itself is still correct, but a record is left behind in the naming tool on every plan until it is fixed.

**`ResourceType value is invalid.`**
No resource type on your instance has that short name. Matching is exact and case-sensitive. List the configured types with `data.proactnaming_resource_types`.

**`Instance value length is invalid. The value must be between N and N characters.`**
`instance` must be exactly the configured width — pad it, for example `001` rather than `1`.

**`You must supply the required components.`**
One or more of `organization`, `location` or `environment` is not a value your naming tool recognises, or a required custom component is missing. Since v0.6.0 the provider checks these while planning, so this arrives as an error against the argument at fault rather than from the API.

**`Unable to Predict Name` … `cannot be evaluated`**
`predict_names_locally` is set and the resource type's validation pattern uses a construct Go cannot evaluate. See the limitations above; unset the setting for that configuration.

**`Generated Name Does Not Match The Plan`**
The name the apply produced differs from the one the plan showed. With `predict_names_locally` set this means the worked-out name is wrong for this deployment, most likely after a naming tool upgrade — unset it and report the difference. Without it, the tool produced different names for the same request, which points at something else writing to it at the same time.

## Concurrency

The Azure Naming Tool stores generated names by reading the whole collection, appending, and writing it back, without locking. Two operations writing at the same time can therefore lose one another's records: a name one caller created can vanish, and a name another deleted can reappear.

In practice this means **avoid running `terraform apply` against the same naming tool concurrently** — from several pipelines, or from a pipeline while someone works in the tool's UI. The provider serialises its own requests within a single run, but it cannot coordinate across separate processes or machines.

## Support

- Provider issues: open an issue in this repository.
- Azure Naming Tool itself: consult your naming tool documentation or its administrator.
- Terraform: [Terraform documentation](https://developer.hashicorp.com/terraform/docs).

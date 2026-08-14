# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Terraform provider (`proactnaming`, registry source `proact-global/proactnaming`) built on `terraform-plugin-framework` that wraps the Azure Naming Tool API to generate standardized Azure resource names. It depends on the sibling library [azurenamingtool-client-go](../azurenamingtool-client-go) (module `github.com/proact-global/azurenamingtool-client-go`) for all HTTP calls to the naming tool — this repo contains no direct HTTP logic, only Terraform schema/CRUD glue around that client.

## Commands

```bash
go build -v ./...              # build
make install                   # fmt, lint, build, install to $GOPATH/bin, generate docs
make fmt                       # gofmt -s -w -e .
make lint                      # golangci-lint run
make generate                  # regenerate docs (tfplugindocs, via tools/ module) — run after schema/description changes
make generate-mappings         # regenerate internal/provider/azurerm_mappings.go from the aztft map.json (needs network)

go test ./...                                          # unit tests (no live instance needed)
go test -run TestAccGenerateNameResource ./internal/provider/  # a single test

# Acceptance tests (create real resources in a live Azure Naming Tool — cost/side effects apply)
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="your-api-key"
export PROACTNAMING_ADMIN_PASSWORD="your-admin-password"  # required for delete/drift-detection tests
TF_ACC=1 go test -v -cover ./internal/provider/
# or:
make testacc
```

`golangci-lint` config is in `.golangci.yml`; `internal/provider/azurerm_mappings.go` is exempted from the `misspell` linter since it's generated content.

### Working against a local build of the client library

Since this provider depends on `azurenamingtool-client-go` as a Go module, use a `go.mod` `replace` directive pointing at the sibling checkout (`../azurenamingtool-client-go`) when developing changes across both repos together, and remove it before committing/releasing.

## Architecture

### Provider wiring (`internal/provider/provider.go`)

`proactnamingProvider.Configure` resolves `host` / `apikey` / `admin_password` from Terraform config, falling back to `PROACTNAMING_HOST` / `PROACTNAMING_APIKEY` / `PROACTNAMING_ADMIN_PASSWORD` env vars, then constructs a single `*azurenamingtool.Client` and hands it to every resource/data source via `resp.DataSourceData` / `resp.ResourceData`. Each resource/data source's `Configure` method type-asserts this into its own `client` field — follow that pattern for any new resource/data source, including the `New*` constructor registered in `provider.go`'s `Resources()`/`DataSources()`.

`admin_password` is technically optional on the client library side, but the provider schema requires it — without it, `terraform destroy` and drift detection cannot work (Admin API-only operations).

### Resource: `proactnaming_generate_name` (`generate_name_resource_source.go`)

This is the only resource, and its lifecycle has non-obvious behavior worth understanding before touching it:

- **All input attributes use `RequiresReplace()`** — there is no in-place update; `Update` is implemented only to return an error if the framework ever calls it unexpectedly.
- **`ModifyPlan` generates a throwaway preview name during `terraform plan`** (calls `GenerateName` then immediately `DeleteName`s it) so the plan shows the real computed name instead of `(known after apply)`. This means every `plan` hits the live naming tool API, not just `apply`.
- **`Read` distinguishes two states**: if `state.ID` is already set, it calls `GetName` to detect drift — `ErrNotFound` from the client triggers `resp.State.RemoveResource(ctx)` so Terraform recreates it. If `ID` is unset (e.g. import or first read), it calls `GenerateName` itself to populate state, since generation is what creates the persistent naming-tool entry.
- **`custom_component` is a repeatable block** (`name`/`value` pairs), merged with the `application` field into a single map via `buildCustomComponents()` before being sent as `GenerateNameRequestCustomComponents` — needed for resource types requiring components beyond the fixed schema fields (e.g. `subnet_tier`, `subnet_instance`).

### Data sources

- `proactnaming_resource_types` (`resourcetypes_data_source.go`) — lists all resource types from `GetResourceTypes()`, converting each to the equivalent snake_case Terraform schema. Each entry's `azurerm_resource_type` is looked up via `getAzurermResourceType(resource, property)`.
- `proactnaming_generated_name` (`generated_name_data_source.go`) — looks up a single previously generated name by numeric `id` via `GetName()`.
- `proactnaming_azurerm_resources` (`azurerm_resources_data_source.go`) — derives a map of `azurerm_*` Terraform resource type → naming-tool short name, built from `GetResourceTypes()` filtered to entries with no `property` qualifier, using the same `getAzurermResourceType` lookup.

### Generated azurerm mapping (`azurerm_mappings.go`)

This file is generated — **do not hand-edit it**. It's produced by `tools/gen_azurerm_mappings.go` (a `//go:build ignore` script, run via `go run tools/gen_azurerm_mappings.go` or `make generate-mappings`) from the [aztft](https://github.com/magodo/aztft) `map.json`, inverting ARM-type → azurerm-resource-name into the lowercase `namespace/resourcetype` → azurerm names lookup used by `getAzurermResourceType`. A scheduled workflow (`.github/workflows/update-azurerm-mappings.yml`, Mondays 08:00 UTC) regenerates it and opens a PR automatically; property-specific override keys (`namespace/resourcetype|property`) are manual additions not touched by regeneration.

### Tests

Unit tests are colocated per file (`*_test.go`). Acceptance tests (`TestAcc*` in `generate_name_resource_test.go`, `generated_name_data_source_test.go`, `azurerm_resources_data_source_test.go`) require `TF_ACC=1` and live credentials, gated by `testAccPreCheck`/`testAccPreCheckWithAdmin` (`provider_test.go`), which skip rather than fail when credentials are absent. Acceptance tests use timestamp-derived instance numbers to avoid naming collisions across runs, and clean up generated entries via the Admin API where relevant (e.g. `testAccDeleteGeneratedNameExternally` simulates out-of-band deletion for drift-detection tests).

### Docs generation

`templates/*.md.tmpl` are the source templates for the registry docs; `make generate` (via `tools/tools.go`'s `tfplugindocs` tool, requires Terraform CLI on PATH) renders them using the live provider schema. CI's `generate` job in `.github/workflows/test.yml` fails the build if regenerating docs produces an uncommitted diff — run `make generate` after any schema, description, or example change.

### Release process

Tag pushes matching `v*` trigger `.github/workflows/release.yml`, which imports a GPG key and runs GoReleaser (`.goreleaser.yml`) to cross-compile for `freebsd`/`windows`/`linux`/`darwin`, sign checksums, and publish to the Terraform Registry using `terraform-registry-manifest.json`.

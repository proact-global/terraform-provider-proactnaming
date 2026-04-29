# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.4.0] - 2026-04-30

### Added
- Drift detection for `proactnaming_generate_name`: the Read function now calls the Admin API with the stored `id` and removes the resource from state when it no longer exists, causing Terraform to plan a recreation on the next apply.
- `proactnaming_generated_name` data source now works correctly — the Admin endpoint returns a raw object rather than a V2 API envelope, which caused every lookup to fail with "get name failed".
- `ErrNotFound` sentinel in the Go client library; callers can use `errors.Is(err, azurenamingtool.ErrNotFound)` to distinguish a missing record from a transient failure.
- New acceptance tests: `TestAccGenerateNameResource_DriftDetection`, `TestAccGenerateNameResource_MultipleResources`, and `TestAccGeneratedNameDataSource`.

### Fixed
- `proactnaming_generated_name` data source: bug where the response from `GET /api/Admin/GetGeneratedName/{id}` was incorrectly unmarshalled as a V2 `ApiResponse` wrapper, always returning "get name failed".
- Mutex lock missing from `GetName` in the Go client.

## [0.3.2] - 2025-12-01

### Fixed
- `azurerm_resource_group` missing from `proactnaming_azurerm_resources` data source output.

## [0.3.1] - 2025-11-28

### Added
- Changelog included in GoReleaser releases.

## [0.3.0] - 2025-11-27

### Added
- `proactnaming_azurerm_resources` data source — returns a map of `azurerm_*` resource types to their short names.
- Automatic regeneration of azurerm mappings via scheduled GitHub Actions workflow.

## [0.2.1] - 2025-11-15

### Fixed
- `azurerm_resource_group` missing from data source output.

## [0.2.0] - 2025-11-01

### Added
- Initial release of Proact Naming Terraform provider.
- `proactnaming_generate_name` resource for generating standardized Azure resource names.
- `proactnaming_resource_types` data source for querying available resource types.
- `proactnaming_generated_name` data source for looking up existing generated names.
- Plan visibility: shows generated names during `terraform plan` instead of "(known after apply)".
- Automatic cleanup of preview entries created during planning.

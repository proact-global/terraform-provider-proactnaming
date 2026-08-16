# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.6.0] - 2026-08-16

### Added
- `predict_names_locally` on the provider. When set, the name shown during `terraform plan` is worked out from the naming tool's configuration instead of being generated and deleted again. Planning then writes nothing to the naming tool and needs no `admin_password`, so a plan-only pipeline no longer requires write credentials. Creating two resources consumed ten records in the naming tool without it, eight of them previews created and deleted, against two with it. It is off by default.
- Requests are checked against the naming tool's configuration while planning. An unknown resource type short name, a value the tool does not accept for a component, or an instance of the wrong width is now reported against the argument at fault, with the accepted values listed and a padded form suggested, rather than arriving from the API phrased in the tool's terms and attributed to nothing.
- `scripts/local-dev.sh` and `examples/local-dev`, for running the provider from the working tree against a naming tool without publishing a release.

### Fixed
- Custom component values are checked the way the naming tool checks them: only when the deployment registers values for that particular component, and never by length. Bounds configured on a custom component are inert in the tool, so enforcing them rejected requests it accepts.

### Changed
- A predicted name that disagrees with the one the tool generates fails the apply with both names reported, rather than leaving Terraform to report an inconsistent result without saying what differed.
- The acceptance suite covers the resource types data source, custom component blocks, the optional `function` argument, the plan-time checks and name prediction, none of which it reached before. It also records every record it creates or replaces in the naming tool and reports, at the end of a run, which of them still exist.

### Internal
- Dependency: `azurenamingtool-client-go` v0.12.0 to v0.13.1, which exposes the naming tool's configuration and can work out a name from it.

## [0.5.0] - 2026-08-15

### Changed
- The provider is now licensed under the MIT License. Earlier releases carried the MPL-2.0 license inherited from HashiCorp's provider scaffolding, which contradicted the MIT license the README has always stated.
- An unusable `host` is now rejected when the provider is configured, with a message naming the value and the form expected. A host given without its protocol previously reached the API layer and failed part way through a plan with Go's `unsupported protocol scheme ""`, which named neither the setting at fault nor the fix.
- A trailing slash on `host` is now trimmed, rather than producing a doubled slash in every request path.

### Fixed
- Drift detection no longer mistakes unrelated failures for a missing name. The provider previously treated any client error whose text contained "not found" — including a 404 from a moved route, a mistyped host, or an intermediate proxy's error page — as evidence the name had been deleted. It removed the resource from state and let Terraform generate a **replacement name for a resource still in use**. Absence is now established from the HTTP status together with the naming tool's own message.
- Failures to remove the preview entry created during `terraform plan` are now reported as a warning naming the entry left behind. The error was previously discarded, and because the naming tool answers an incorrect admin password with HTTP 200 and a "FAILURE" body rather than an authentication error, a mistyped `admin_password` produced healthy-looking plans while leaking one record per resource on every plan.
- `terraform destroy` no longer fails against a naming tool holding no generated names, which is reported with a pluralised "Generated Names not found!" that the client did not recognise.
- Configuration errors from client creation are now attributed to the `host` attribute, so Terraform points at the argument that needs changing.

### Internal
- Dependency: `azurenamingtool-client-go` v0.10.0 to v0.12.0.
- Unit tests now run in CI on every push and pull request. Acceptance tests are opt-in through the `ACCEPTANCE_TESTS_ENABLED` repository variable, and fail rather than skip when credentials are absent in CI, where a skipped suite is indistinguishable from a passing one.
- Acceptance test component values (organisation, resource types, location, environment, instance width) are configurable through `PROACTNAMING_TEST_*` variables, so the suite can run against any naming tool rather than only the one it was written against.
- Acceptance runs are serialised. The naming tool stores generated names by reading the whole collection, appending, and writing it back without locking, so concurrent runs silently lose each other's records.

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

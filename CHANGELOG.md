# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Initial release of Proact Naming Terraform provider
- `proactnaming_generate_name` resource for generating standardized Azure resource names
- `proactnaming_resourcetypes` data source for querying available resource types
- `proactnaming_generated_name` data source for looking up existing generated names
- Support for Azure Naming Tool API integration
- Plan visibility showing generated names before apply
- Automatic cleanup of preview entries during planning

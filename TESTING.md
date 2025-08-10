# Testing

## Prerequisites

Set up your environment variables:

```bash
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="your-api-key"
export PROACTNAMING_ADMIN_PASSWORD="your-admin-password"  # Optional
```

## Running Tests

### Unit Tests
```bash
go test ./internal/provider/
```

### Acceptance Tests
```bash
export TF_ACC=1
go test -v ./internal/provider/ -run TestAcc
```

### Manual Testing
```bash
cd examples/basic
terraform init
terraform plan
terraform apply
```

## Notes

- The provider automatically cleans up preview entries during planning
- Each resource maintains exactly one entry in the Azure Naming Tool during its lifecycle
- Use unique `instance` values when testing to avoid naming conflicts

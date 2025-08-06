# Testing Notes

## Current Limitations

### Delete Functionality
The Azure Naming Tool API currently doesn't support deletion via API key authentication. Delete operations require both an API key and a plain text password, which is not secure and not supported by this provider.

**Impact on Testing:**
- Resources created during tests remain in the naming tool
- The same name configuration cannot be recreated
- Acceptance tests use timestamp-based unique identifiers to avoid conflicts

### Acceptance Tests

To run acceptance tests:

```bash
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="your-api-key"
export TF_ACC=1

go test -v ./internal/provider/ -run TestAcc
```

**Note:** Each test run will create new resources in the naming tool that cannot be automatically cleaned up.

## Manual Testing

For manual testing without accumulating test resources:

```bash
cd examples/generatename
terraform init
terraform plan
terraform apply
```

Use unique values for `instance` field in each run to avoid conflicts.

## Future Enhancements

Once the API supports delete operations with API key authentication:
1. Enable full CRUD testing
2. Add comprehensive test coverage
3. Remove timestamp-based workarounds

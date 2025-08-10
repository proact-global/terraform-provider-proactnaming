# Provider Configuration Examples

This directory contains examples of different ways to configure the Proact Naming provider.

## Configuration Methods

### 1. Direct Configuration (`direct-config.tf`)
Configure the provider directly in your Terraform configuration with hardcoded values.

### 2. Environment Variables (`env-vars.tf`)
Configure the provider using environment variables for better security.

### 3. Mixed Approach (`mixed-config.tf`)
Combine direct configuration with environment variable fallbacks.

## Security Best Practices

1. **Never commit API keys**: Use environment variables or encrypted storage
2. **Use .gitignore**: Ensure `*.tfvars` files are excluded from version control
3. **Rotate keys regularly**: Update API keys according to your security policy
4. **Limit permissions**: Use API keys with minimal required permissions

## Environment Variables

Set these environment variables for secure configuration:

```bash
export PROACTNAMING_HOST="https://your-naming-tool.azurewebsites.net"
export PROACTNAMING_APIKEY="your-api-key"
export PROACTNAMING_ADMIN_PASSWORD="your-admin-password"  # Optional
```

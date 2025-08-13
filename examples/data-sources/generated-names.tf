# Copyright (c) HashiCorp, Inc.

# Example of querying specific generated names by ID
# Note: You need actual IDs from your Azure Naming Tool to use this data source

# Uncomment and replace with real ID when testing:
# data "proactnaming_generated_name" "existing_resource" {
#   id = "123"  # Replace with actual ID from your naming tool
# }

# Example output usage (uncomment when you have real IDs):
# output "existing_name_details" {
#   description = "Details of an existing generated name"
#   value = {
#     id   = data.proactnaming_generated_name.existing_resource.id
#     name = data.proactnaming_generated_name.existing_resource.generated_name[0].resource_name
#     type = data.proactnaming_generated_name.existing_resource.generated_name[0].resource_type_name
#   }
# }

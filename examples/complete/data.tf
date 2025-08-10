# Query available resource types from the Azure Naming Tool
data "proactnaming_resourcetypes" "available" {}

# Example of querying a specific generated name (if you have an existing ID)
# data "proactnaming_generated_name" "existing" {
#   id = "123"
# }

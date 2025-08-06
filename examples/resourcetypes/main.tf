terraform {
  required_providers {
    proactnaming = {
      source = "hashicorp.com/edu/proactnaming"
    }
  }
}

provider "proactnaming" {
  host   = "https://mangopato-namingtool.azurewebsites.net"
  apikey = "6f510d50-ba62-40e1-a432-c3fc0a530483"
}

data "proactnaming_resource_types" "edu" {}

data "proactnaming_generated_name" "example" {
  id = "12"
}

output "name" {
  value = data.proactnaming_generated_name.example
}

locals {
  output_typesname = [
    for k in data.proactnaming_resource_types.edu.resource_types : k.short_name
  ]
}

# output "types" {
#   value = data.proactnaming_resource_types.edu.resource_types #local.output_typesname
# }
terraform {
  required_providers {
    proactnaming = {
      source = "hashicorp.com/edu/proactnaming"
    }
  }
}

provider "proactnaming" {
  host     = "https://mangopato-namingtool.azurewebsites.net"
  apikey = "6f510d50-ba62-40e1-a432-c3fc0a530483"
    }

data "proactnaming_resource_types" "edu" {}

output "resource_types" {
  value = data.proactnaming_resource_types.edu
}

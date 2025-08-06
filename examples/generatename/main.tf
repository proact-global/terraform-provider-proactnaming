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

resource "proactnaming_generate_name" "name" {
  organization  = "man"
  resource_type = "rg"
  application   = "app"
  function      = "test"
  instance      = "004"
  location      = "euw"
  environment   = "dev"

}

data "proactnaming_generated_name" "example" {
  id = "31"
}

output "proactnaming_generate_name" {
  value = data.proactnaming_generated_name.example
}
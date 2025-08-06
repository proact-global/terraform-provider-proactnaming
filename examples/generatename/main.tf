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
  resource_type = "st"
  application   = "app"
  function      = ""
  instance      = "002"
  location      = "euw"
  environment   = "dev"

}

output "proactnaming_generate_name_id" {
  value = proactnaming_generate_name.name
}
terraform {
  required_providers {
    proactnaming = {
      source = "hashicorp.com/edu/proactnaming"
    }
  }
}

provider "proactnaming" {
  host     = "http://localhost:19090"
  username = "user1"
  password = "test123"
}

data "proactnaming_coffees" "edu" {}

output "coffees" {
  value = data.proactnaming_coffees.edu
}

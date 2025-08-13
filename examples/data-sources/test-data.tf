# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    proactnaming = {
      source = "proact-global/proactnaming"
    }
  }
}

provider "proactnaming" {
  host           = "https://test.example.com"
  apikey         = "test-key"
  admin_password = "test-password"
}

# Test data source
data "proactnaming_resource_types" "test" {}

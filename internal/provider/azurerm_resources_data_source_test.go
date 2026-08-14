// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAzurermResourcesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAzurermResourcesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The map must be populated.
					resource.TestCheckResourceAttrSet(
						"data.proactnaming_azurerm_resources.test", "resources.%"),
					// Well-known azurerm types must be present.
					resource.TestCheckResourceAttrSet(
						"data.proactnaming_azurerm_resources.test", "resources.azurerm_storage_account"),
					resource.TestCheckResourceAttrSet(
						"data.proactnaming_azurerm_resources.test", "resources.azurerm_resource_group"),
					// Short names must be non-empty strings (letters only).
					resource.TestMatchResourceAttr(
						"data.proactnaming_azurerm_resources.test", "resources.azurerm_storage_account",
						regexp.MustCompile(`^[a-z]+$`)),
					resource.TestMatchResourceAttr(
						"data.proactnaming_azurerm_resources.test", "resources.azurerm_resource_group",
						regexp.MustCompile(`^[a-z]+$`)),
				),
			},
		},
	})
}

func testAccAzurermResourcesDataSourceConfig() string {
	return `
provider "proactnaming" {
  # Uses PROACTNAMING_HOST and PROACTNAMING_APIKEY environment variables
}

data "proactnaming_azurerm_resources" "test" {}
`
}

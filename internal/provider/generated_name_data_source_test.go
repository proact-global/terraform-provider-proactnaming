// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGeneratedNameDataSource verifies the proactnaming_generated_name data source can
// look up a name that was created by the proactnaming_generate_name resource using its ID.
func TestAccGeneratedNameDataSource(t *testing.T) {
	org, rt := testAccOrg(), testAccResourceType()
	loc, env := testAccLocation(), testAccEnvironment()
	instance := testAccInstance(t, 0)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGeneratedNameDataSourceConfig(org, rt, instance, loc, env),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source must have a non-empty generated_name list.
					resource.TestCheckResourceAttr(
						"data.proactnaming_generated_name.lookup", "generated_name.#", "1"),
					// The looked-up resource_name must match what the resource generated.
					resource.TestCheckResourceAttrPair(
						"data.proactnaming_generated_name.lookup", "generated_name.0.resource_name",
						"proactnaming_generate_name.source", "resource_name"),
					// The looked-up ID must match the resource ID.
					resource.TestCheckResourceAttrPair(
						"data.proactnaming_generated_name.lookup", "id",
						"proactnaming_generate_name.source", "id"),
					// resource_type_name must be non-empty.
					resource.TestCheckResourceAttrSet(
						"data.proactnaming_generated_name.lookup", "generated_name.0.resource_type_name"),
				),
			},
		},
	})
}

func testAccGeneratedNameDataSourceConfig(organization, resourceType, instance, location, environment string) string {
	return fmt.Sprintf(`
provider "proactnaming" {
  # Uses PROACTNAMING_HOST, PROACTNAMING_APIKEY, and PROACTNAMING_ADMIN_PASSWORD environment variables
}

resource "proactnaming_generate_name" "source" {
  organization  = %[1]q
  resource_type = %[2]q
  application   = "dstest"
  instance      = %[3]q
  location      = %[4]q
  environment   = %[5]q
}

data "proactnaming_generated_name" "lookup" {
  id = proactnaming_generate_name.source.id
}
`, organization, resourceType, instance, location, environment)
}

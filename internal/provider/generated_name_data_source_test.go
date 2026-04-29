// Copyright (c) HashiCorp, Inc.

package provider

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGeneratedNameDataSource verifies the proactnaming_generated_name data source can
// look up a name that was created by the proactnaming_generate_name resource using its ID.
func TestAccGeneratedNameDataSource(t *testing.T) {
	timestamp := time.Now().Unix()
	instance := strconv.FormatInt(timestamp%1000, 10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGeneratedNameDataSourceConfig(instance),
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

func testAccGeneratedNameDataSourceConfig(instance string) string {
	return fmt.Sprintf(`
provider "proactnaming" {
  # Uses PROACTNAMING_HOST, PROACTNAMING_APIKEY, and PROACTNAMING_ADMIN_PASSWORD environment variables
}

resource "proactnaming_generate_name" "source" {
  organization  = "man"
  resource_type = "st"
  application   = "dstest"
  instance      = %[1]q
  location      = "euw"
  environment   = "dev"
}

data "proactnaming_generated_name" "lookup" {
  id = proactnaming_generate_name.source.id
}
`, instance)
}

package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGenerateNameResource(t *testing.T) {
	// Generate unique instance number using timestamp to avoid conflicts
	timestamp := time.Now().Unix()
	uniqueInstance := strconv.FormatInt(timestamp%1000, 10) // Use last 3 digits

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGenerateNameResourceConfig("man", "st", "webapp", "test", uniqueInstance, "euw", "dev"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "organization", "man"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "resource_type", "st"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "application", "webapp"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "function", "test"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "instance", uniqueInstance),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "location", "euw"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "environment", "dev"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "id"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "resource_name"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "success", "true"),
					// Verify the generated name follows expected pattern
					resource.TestMatchResourceAttr("proactnaming_generate_name.test", "resource_name",
						regexp.MustCompile(`^man-st-webapp.*-euw-dev$`)),
				),
			},
			// Test replacement behavior by changing instance
			{
				Config: testAccGenerateNameResourceConfig("man", "st", "webapp", "test", "999", "euw", "dev"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "instance", "999"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "resource_name"),
					// Verify the new name is different and follows pattern
					resource.TestMatchResourceAttr("proactnaming_generate_name.test", "resource_name",
						regexp.MustCompile(`^man-st-webapp.*999.*-euw-dev$`)),
				),
			},
		},
	})
}

// TestAccGenerateNameResource_UniqueNames tests that different configurations generate different names
func TestAccGenerateNameResource_UniqueNames(t *testing.T) {
	// Skip this test as it would create multiple resources without cleanup
	t.Skip("Skipping multiple resource test until delete functionality is available. " +
		"This test would create multiple resources in the naming tool without ability to clean them up.")
}

func testAccGenerateNameResourceConfig(organization, resourceType, application, function, instance, location, environment string) string {
	return fmt.Sprintf(`
provider "proactnaming" {
  # Uses PROACTNAMING_HOST and PROACTNAMING_APIKEY environment variables
}

resource "proactnaming_generate_name" "test" {
  organization  = %[1]q
  resource_type = %[2]q
  application   = %[3]q
  function      = %[4]q
  instance      = %[5]q
  location      = %[6]q
  environment   = %[7]q
}
`, organization, resourceType, application, function, instance, location, environment)
}

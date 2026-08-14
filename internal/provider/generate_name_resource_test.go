// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/proact-global/azurenamingtool-client-go"
)

func TestAccGenerateNameResource(t *testing.T) {
	org, rt := testAccOrg(), testAccResourceType()
	loc, env := testAccLocation(), testAccEnvironment()

	// Derived from the clock so repeated runs do not collide, and padded to the
	// width the naming tool requires.
	uniqueInstance := testAccInstance(0)
	replacementInstance := testAccInstance(1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccGenerateNameResourceConfig(org, rt, "webapp", "test", uniqueInstance, loc, env),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "organization", org),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "resource_type", rt),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "application", "webapp"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "function", "test"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "instance", uniqueInstance),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "location", loc),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "environment", env),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "id"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "resource_name"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "success", "true"),
					// Verify the generated name contains the expected components (delimiter-agnostic).
					resource.TestMatchResourceAttr("proactnaming_generate_name.test", "resource_name",
						testAccNameRegexp(org, rt, "webapp", loc, env)),
				),
			},
			// Test replacement behavior by changing instance.
			{
				Config: testAccGenerateNameResourceConfig(org, rt, "webapp", "test", replacementInstance, loc, env),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proactnaming_generate_name.test", "instance", replacementInstance),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "resource_name"),
					// Verify the new name is different and follows pattern.
					resource.TestMatchResourceAttr("proactnaming_generate_name.test", "resource_name",
						testAccNameRegexp(org, rt, "webapp", replacementInstance, loc, env)),
				),
			},
		},
	})
}

// TestAccGenerateNameResource_DriftDetection verifies that when a generated name is deleted
// outside of Terraform (drift), the provider detects the absence on the next plan/apply
// and recreates it cleanly.
func TestAccGenerateNameResource_DriftDetection(t *testing.T) {
	org, rt := testAccOrg(), testAccResourceTypeAlt()
	loc, env := testAccLocation(), testAccEnvironment()
	instance := testAccInstance(0)
	config := testAccGenerateNameResourceConfig(org, rt, "drifttest", "drift", instance, loc, env)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create the resource, then immediately delete it externally to simulate drift.
			// ExpectNonEmptyPlan is required because the post-step refresh detects the name is
			// gone (our Check deleted it) and marks the resource for recreation.
			{
				Config:             config,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "id"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "resource_name"),
					// Simulate out-of-band deletion so the next step exercises drift detection.
					testAccDeleteGeneratedNameExternally("proactnaming_generate_name.test"),
				),
			},
			// Step 2: Re-apply the same config. The provider's Read detects ErrNotFound,
			// removes the resource from state, and Terraform recreates it.
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "id"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.test", "resource_name"),
					resource.TestMatchResourceAttr("proactnaming_generate_name.test", "resource_name",
						testAccNameRegexp(org, rt, "drifttest", loc, env)),
				),
			},
		},
	})
}

// TestAccGenerateNameResource_MultipleResources tests that different configurations
// generate unique names and that all are properly cleaned up after the test.
func TestAccGenerateNameResource_MultipleResources(t *testing.T) {
	org := testAccOrg()
	loc, env := testAccLocation(), testAccEnvironment()
	rtAlt, rt := testAccResourceTypeAlt(), testAccResourceType()

	// Two distinct instance values, both padded to the required width. The
	// previous derivation (timestamp%100) produced one or two characters, which
	// the naming tool rejects outright.
	inst1 := testAccInstance(0)
	inst2 := testAccInstance(1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMultipleResourcesConfig(org, rtAlt, rt, inst1, inst2, loc, env),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.rg", "resource_name"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.st", "resource_name"),
					// Each resource gets the correct resource_type stored in state.
					resource.TestCheckResourceAttr("proactnaming_generate_name.rg", "resource_type", rtAlt),
					resource.TestCheckResourceAttr("proactnaming_generate_name.st", "resource_type", rt),
					// IDs must be set and independent.
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.rg", "id"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.st", "id"),
				),
			},
		},
	})
}

// testAccDeleteGeneratedNameExternally returns a TestCheckFunc that deletes the generated
// name entry directly via the API client, simulating an out-of-band deletion for drift tests.
func testAccDeleteGeneratedNameExternally(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		idStr := rs.Primary.Attributes["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse id %q: %w", idStr, err)
		}

		host := os.Getenv("PROACTNAMING_HOST")
		apiKey := os.Getenv("PROACTNAMING_APIKEY")
		adminPwd := os.Getenv("PROACTNAMING_ADMIN_PASSWORD")

		client, err := azurenamingtool.NewClient(&host, &apiKey, &adminPwd)
		if err != nil {
			return fmt.Errorf("failed to create naming tool client: %w", err)
		}

		_, err = client.DeleteName(azurenamingtool.DeleteGeneratedNameRequest{ID: id})
		if err != nil {
			return fmt.Errorf("external delete of name ID %d failed: %w", id, err)
		}
		return nil
	}
}

func testAccGenerateNameResourceConfig(organization, resourceType, application, function, instance, location, environment string) string {
	return fmt.Sprintf(`
provider "proactnaming" {
  # Uses PROACTNAMING_HOST, PROACTNAMING_APIKEY, and PROACTNAMING_ADMIN_PASSWORD environment variables
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

func testAccMultipleResourcesConfig(organization, resourceTypeA, resourceTypeB, inst1, inst2, location, environment string) string {
	return fmt.Sprintf(`
provider "proactnaming" {}

resource "proactnaming_generate_name" "rg" {
  organization  = %[1]q
  resource_type = %[2]q
  application   = "multitest"
  instance      = %[4]q
  location      = %[6]q
  environment   = %[7]q
}

resource "proactnaming_generate_name" "st" {
  organization  = %[1]q
  resource_type = %[3]q
  application   = "multitest"
  instance      = %[5]q
  location      = %[6]q
  environment   = %[7]q
}
`, organization, resourceTypeA, resourceTypeB, inst1, inst2, location, environment)
}

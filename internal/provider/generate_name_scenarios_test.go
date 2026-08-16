// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Scenarios the original acceptance suite did not reach: the resource types data
// source, custom component blocks, the checks performed while planning, and
// predicting the name instead of generating one.

// TestAccResourceTypesDataSource covers proactnaming_resource_types, which had
// no acceptance test at all despite being one of the provider's four surfaces.
func TestAccResourceTypesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "proactnaming" {}

data "proactnaming_resource_types" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.proactnaming_resource_types.all", "resource_types.#"),
					resource.TestCheckResourceAttrSet("data.proactnaming_resource_types.all", "resource_types.0.resource"),
					resource.TestCheckResourceAttrSet("data.proactnaming_resource_types.all", "resource_types.0.short_name"),
				),
			},
		},
	})
}

// TestAccGenerateNameResource_CustomComponents covers the custom_component
// block, a whole part of the schema the suite never exercised against a live
// naming tool.
//
// It is skipped unless a custom component is named, because which components a
// deployment defines is its own business and there is no safe default.
func TestAccGenerateNameResource_CustomComponents(t *testing.T) {
	name := envOrDefault("PROACTNAMING_TEST_CUSTOM_COMPONENT", "")
	value := envOrDefault("PROACTNAMING_TEST_CUSTOM_COMPONENT_VALUE", "")
	if name == "" || value == "" {
		t.Skip("set PROACTNAMING_TEST_CUSTOM_COMPONENT and PROACTNAMING_TEST_CUSTOM_COMPONENT_VALUE " +
			"to a custom component this naming tool defines, to exercise custom_component blocks")
	}

	org, rt := testAccOrg(), testAccResourceType()
	loc, env := testAccLocation(), testAccEnvironment()
	instance := testAccInstance(t, 0)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "proactnaming" {}

resource "proactnaming_generate_name" "custom" {
  organization  = %[1]q
  resource_type = %[2]q
  application   = "cctest"
  instance      = %[3]q
  location      = %[4]q
  environment   = %[5]q

  custom_component {
    name  = %[6]q
    value = %[7]q
  }
}
`, org, rt, instance, loc, env, name, value),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.custom", "resource_name"),
					resource.TestCheckResourceAttr("proactnaming_generate_name.custom", "custom_component.0.name", name),
					resource.TestCheckResourceAttr("proactnaming_generate_name.custom", "custom_component.0.value", value),
					// The custom value has to appear in the generated name.
					resource.TestMatchResourceAttr("proactnaming_generate_name.custom", "resource_name",
						regexp.MustCompile(regexp.QuoteMeta(value))),
					testAccRecord(t, "proactnaming_generate_name.custom", ledgerCreated),
				),
			},
		},
	})
}

// TestAccGenerateNameResource_PlanTimeChecksRejectBadInput is the acceptance
// counterpart to the validation unit tests: it confirms that against a real
// naming tool the request is refused while planning, with the argument named,
// rather than being sent and refused by the API.
//
// It also confirms nothing is created, since the plan never completes.
func TestAccGenerateNameResource_PlanTimeChecksRejectBadInput(t *testing.T) {
	org := testAccOrg()
	loc, env := testAccLocation(), testAccEnvironment()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// A short name no deployment will have.
				Config: fmt.Sprintf(`
provider "proactnaming" {}

resource "proactnaming_generate_name" "bad" {
  organization  = %[1]q
  resource_type = "zzzznotatype"
  application   = "badtest"
  instance      = "001"
  location      = %[2]q
  environment   = %[3]q
}
`, org, loc, env),
				ExpectError: regexp.MustCompile(`(?s)Unknown Resource Type|no resource type with the short name`),
			},
			{
				// An instance that is numeric but the wrong width.
				Config: fmt.Sprintf(`
provider "proactnaming" {}

resource "proactnaming_generate_name" "bad" {
  organization  = %[1]q
  resource_type = %[2]q
  application   = "badtest"
  instance      = "1"
  location      = %[3]q
  environment   = %[4]q
}
`, org, testAccResourceType(), loc, env),
				ExpectError: regexp.MustCompile(`(?s)Invalid Length for instance|must be exactly`),
			},
		},
	})
}

// TestAccGenerateNameResource_PredictedNameMatchesGenerated is the test that
// justifies predict_names_locally existing.
//
// With prediction enabled the plan shows a name the provider worked out itself,
// and the apply stores the name the naming tool generates. If the reproduction
// of the tool's algorithm were wrong for this deployment, the two would differ
// and the apply would fail. Passing means the provider's understanding of how
// this deployment builds a name agrees with the deployment.
//
// It is the safety net for the whole feature, and the check to run again after
// the naming tool is upgraded.
func TestAccGenerateNameResource_PredictedNameMatchesGenerated(t *testing.T) {
	org, rt := testAccOrg(), testAccResourceType()
	loc, env := testAccLocation(), testAccEnvironment()
	instance := testAccInstance(t, 0)

	config := fmt.Sprintf(`
provider "proactnaming" {
  predict_names_locally = true
}

resource "proactnaming_generate_name" "predicted" {
  organization  = %[1]q
  resource_type = %[2]q
  application   = "predict"
  instance      = %[3]q
  location      = %[4]q
  environment   = %[5]q
}
`, org, rt, instance, loc, env)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.predicted", "id"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.predicted", "resource_name"),
					resource.TestMatchResourceAttr("proactnaming_generate_name.predicted", "resource_name",
						testAccNameRegexp(org, rt, "predict", instance, loc, env)),
					testAccRecord(t, "proactnaming_generate_name.predicted", ledgerCreated),
				),
			},
			{
				// A second plan over unchanged configuration must be empty. With
				// prediction on this exercises the path again without applying,
				// and confirms the predicted name is stable.
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// TestAccGenerateNameResource_OptionalFunctionOmitted covers leaving out the one
// optional argument, which the suite never did.
func TestAccGenerateNameResource_OptionalFunctionOmitted(t *testing.T) {
	org, rt := testAccOrg(), testAccResourceType()
	loc, env := testAccLocation(), testAccEnvironment()
	instance := testAccInstance(t, 0)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWithAdmin(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "proactnaming" {}

resource "proactnaming_generate_name" "nofunction" {
  organization  = %[1]q
  resource_type = %[2]q
  application   = "nofunc"
  instance      = %[3]q
  location      = %[4]q
  environment   = %[5]q
}
`, org, rt, instance, loc, env),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("proactnaming_generate_name.nofunction", "function"),
					resource.TestCheckResourceAttrSet("proactnaming_generate_name.nofunction", "resource_name"),
					testAccRecord(t, "proactnaming_generate_name.nofunction", ledgerCreated),
				),
			},
		},
	})
}

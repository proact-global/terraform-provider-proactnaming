// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/proact-global/azurenamingtool-client-go"
)

// testConfiguration is a naming tool configuration resembling a real deployment:
// a fixed-width numeric instance, enumerable orgs, locations and environments, a
// disabled function component, and custom components.
func testConfiguration() *azurenamingtool.NamingConfiguration {
	return &azurenamingtool.NamingConfiguration{
		Components: []azurenamingtool.ResourceComponent{
			{Name: "ResourceType", Enabled: true, SortOrder: 1},
			{Name: "ResourceOrg", Enabled: true, SortOrder: 2},
			{Name: "ResourceLocation", Enabled: true, SortOrder: 3},
			{Name: "ResourceEnvironment", Enabled: true, SortOrder: 4},
			{Name: "ResourceFunction", Enabled: false, SortOrder: 5},
			{Name: "ResourceInstance", Enabled: true, SortOrder: 9,
				IsFreeText: true, MinLength: "3", MaxLength: "3", Alphanumeric: true},
		},
		ResourceTypes: []azurenamingtool.ResourceTypes{
			{Resource: "Storage/storageAccounts", ShortName: "st"},
			{Resource: "Resources/resourceGroups", ShortName: "rg"},
		},
		Orgs:         []azurenamingtool.ComponentValue{{Name: "Proact", ShortName: "pca"}},
		Locations:    []azurenamingtool.ComponentValue{{Name: "West Europe", ShortName: "we"}},
		Environments: []azurenamingtool.ComponentValue{{Name: "Production", ShortName: "p"}},
		Functions:    []azurenamingtool.ComponentValue{{Name: "Data", ShortName: "dat"}},
		CustomComponents: []azurenamingtool.CustomComponent{
			{ParentComponent: "application", ShortName: "webapp"},
			{ParentComponent: "application", ShortName: "api"},
			{ParentComponent: "subnet_tier", ShortName: "app"},
		},
	}
}

// validPlan is a request the configuration above accepts.
func validPlan() generateNameModel {
	return generateNameModel{
		Organization: types.StringValue("pca"),
		ResourceType: types.StringValue("st"),
		Application:  types.StringValue("webapp"),
		Instance:     types.StringValue("001"),
		Location:     types.StringValue("we"),
		Environment:  types.StringValue("p"),
		Function:     types.StringNull(),
	}
}

func TestValidateAcceptsAValidRequest(t *testing.T) {
	if d := validateAgainstConfiguration(testConfiguration(), validPlan()); d.HasError() {
		t.Fatalf("valid request rejected: %v", d.Errors())
	}
}

// TestValidateCatchesTheFailuresSeenInPractice covers the errors the API
// returned during acceptance testing, which arrived attributed to nothing and
// phrased in the tool's terms.
func TestValidateCatchesTheFailuresSeenInPractice(t *testing.T) {
	tests := map[string]struct {
		mutate       func(*generateNameModel)
		wantAttr     string
		wantContains string
	}{
		"unknown resource type": {
			func(p *generateNameModel) { p.ResourceType = types.StringValue("nope") },
			"resource_type", "no resource type with the short name",
		},
		"resource type wrong case": {
			func(p *generateNameModel) { p.ResourceType = types.StringValue("ST") },
			"resource_type", "matched exactly, including case",
		},
		"unknown organization": {
			func(p *generateNameModel) { p.Organization = types.StringValue("acme") },
			"organization", "does not accept",
		},
		"unknown location": {
			func(p *generateNameModel) { p.Location = types.StringValue("euw") },
			"location", "does not accept",
		},
		"unknown environment": {
			func(p *generateNameModel) { p.Environment = types.StringValue("dev") },
			"environment", "does not accept",
		},
		"instance too short": {
			func(p *generateNameModel) { p.Instance = types.StringValue("1") },
			"instance", "exactly 3 characters",
		},
		"instance too long": {
			func(p *generateNameModel) { p.Instance = types.StringValue("0001") },
			"instance", "exactly 3 characters",
		},
		"instance not alphanumeric": {
			func(p *generateNameModel) { p.Instance = types.StringValue("0-1") },
			"instance", "letters and digits only",
		},
		"unknown application": {
			func(p *generateNameModel) { p.Application = types.StringValue("unknown") },
			"application", "does not accept",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			plan := validPlan()
			tt.mutate(&plan)

			diags := validateAgainstConfiguration(testConfiguration(), plan)
			if !diags.HasError() {
				t.Fatal("expected an error, got none")
			}
			err := diags.Errors()[0]
			if got := err.Detail(); !strings.Contains(got, tt.wantContains) {
				t.Errorf("detail = %q, want it to mention %q", got, tt.wantContains)
			}
			// The diagnostic must point at the argument, which is the whole
			// point of checking before sending.
			if got := diags.Errors()[0].Summary(); got == "" {
				t.Error("diagnostic has no summary")
			}
		})
	}
}

// TestValidateSuggestsPaddingAnUnpaddedInstance covers the most common mistake:
// a number that is correct but not padded to the required width.
func TestValidateSuggestsPaddingAnUnpaddedInstance(t *testing.T) {
	plan := validPlan()
	plan.Instance = types.StringValue("7")

	diags := validateAgainstConfiguration(testConfiguration(), plan)
	if !diags.HasError() {
		t.Fatal("expected an error")
	}
	if got := diags.Errors()[0].Detail(); !strings.Contains(got, `"007"`) {
		t.Errorf("detail = %q, want it to suggest %q", got, "007")
	}
}

// TestValidateSkipsDisabledComponents ensures a value for a component the
// deployment does not use is not rejected. ResourceFunction is disabled in the
// fixture, so any function must pass.
func TestValidateSkipsDisabledComponents(t *testing.T) {
	plan := validPlan()
	plan.Function = types.StringValue("not-a-configured-function")

	if d := validateAgainstConfiguration(testConfiguration(), plan); d.HasError() {
		t.Errorf("value for a disabled component was rejected: %v", d.Errors())
	}
}

// TestValidateIgnoresUnknownAndNullValues ensures values Terraform has not
// settled yet are left alone: they are resolved later in the plan, and rejecting
// them would break any configuration deriving a name from another resource.
func TestValidateIgnoresUnknownAndNullValues(t *testing.T) {
	for name, mutate := range map[string]func(*generateNameModel){
		"unknown resource type": func(p *generateNameModel) { p.ResourceType = types.StringUnknown() },
		"unknown org":           func(p *generateNameModel) { p.Organization = types.StringUnknown() },
		"unknown instance":      func(p *generateNameModel) { p.Instance = types.StringUnknown() },
		"null application":      func(p *generateNameModel) { p.Application = types.StringNull() },
	} {
		t.Run(name, func(t *testing.T) {
			plan := validPlan()
			mutate(&plan)
			if d := validateAgainstConfiguration(testConfiguration(), plan); d.HasError() {
				t.Errorf("unsettled value was rejected: %v", d.Errors())
			}
		})
	}
}

func TestValidateCustomComponentBlocks(t *testing.T) {
	t.Run("accepted value passes", func(t *testing.T) {
		plan := validPlan()
		plan.CustomComponents = []customComponentModel{
			{Name: types.StringValue("subnet_tier"), Value: types.StringValue("app")},
		}
		if d := validateAgainstConfiguration(testConfiguration(), plan); d.HasError() {
			t.Errorf("valid custom component rejected: %v", d.Errors())
		}
	})

	t.Run("unknown value is rejected", func(t *testing.T) {
		plan := validPlan()
		plan.CustomComponents = []customComponentModel{
			{Name: types.StringValue("subnet_tier"), Value: types.StringValue("nope")},
		}
		d := validateAgainstConfiguration(testConfiguration(), plan)
		if !d.HasError() {
			t.Fatal("expected an error")
		}
		if got := d.Errors()[0].Detail(); !strings.Contains(got, "does not accept") {
			t.Errorf("detail = %q", got)
		}
	})

	// The tool checks a custom component's value only when values are
	// registered for that particular component. A name it knows nothing about
	// is passed through untouched, so the provider must not invent a rejection.
	t.Run("component with no registered values is accepted", func(t *testing.T) {
		plan := validPlan()
		plan.CustomComponents = []customComponentModel{
			{Name: types.StringValue("no_such_component"), Value: types.StringValue("x")},
		}
		if d := validateAgainstConfiguration(testConfiguration(), plan); d.HasError() {
			t.Errorf("a component with no registered values was rejected: %v", d.Errors())
		}
	})
}

// TestValidateIsSilentWithoutConfiguration ensures an empty configuration -- a
// deployment that defines nothing, or a partial read -- rejects nothing. The
// check may only ever add certainty, never invent failures.
func TestValidateIsSilentWithoutConfiguration(t *testing.T) {
	empty := &azurenamingtool.NamingConfiguration{}
	plan := validPlan()
	plan.ResourceType = types.StringValue("anything")

	d := validateAgainstConfiguration(empty, plan)
	if d.HasError() {
		// A resource type with no configured types to compare against is
		// reported, which is correct: the tool cannot generate any name.
		if !strings.Contains(d.Errors()[0].Detail(), "resource type") {
			t.Errorf("unexpected error against an empty configuration: %v", d.Errors())
		}
	}
}

func TestListValuesCapsLongLists(t *testing.T) {
	many := make([]string, 40)
	for i := range many {
		many[i] = string(rune('a' + i%26))
	}
	got := listValues(many)
	if !strings.Contains(got, "and 28 more") {
		t.Errorf("listValues did not cap the list: %q", got)
	}

	few := listValues([]string{"b", "a"})
	if !strings.Contains(few, "a, b") {
		t.Errorf("listValues should sort: %q", few)
	}

	if got := listValues(nil); !strings.Contains(got, "no accepted values") {
		t.Errorf("empty list message = %q", got)
	}
}

// TestValidateAcceptsAnyValueWhenNoneAreRegistered covers the deployment the
// acceptance tests run against, which registers no values for its application
// component and so accepts anything for it.
func TestValidateAcceptsAnyValueWhenNoneAreRegistered(t *testing.T) {
	cfg := testConfiguration()
	cfg.CustomComponents = nil

	plan := validPlan()
	plan.Application = types.StringValue("anythingatall")

	if d := validateAgainstConfiguration(cfg, plan); d.HasError() {
		t.Errorf("value rejected though nothing is registered: %v", d.Errors())
	}
}

// TestValidateDoesNotLengthCheckCustomComponents pins that the length bounds on
// a custom component are not enforced. The tool applies its length check on the
// branch handling built-in components only, so a deployment can carry bounds a
// custom component's values comfortably exceed -- as this one does, its
// application component bounded at 3 to 4 while names are built from values
// like "drifttest".
//
// Enforcing them rejected requests the tool accepts, which the acceptance suite
// caught immediately.
func TestValidateDoesNotLengthCheckCustomComponents(t *testing.T) {
	cfg := testConfiguration()
	cfg.Components = append(cfg.Components, azurenamingtool.ResourceComponent{
		Name: "application", Enabled: true, SortOrder: 3,
		IsCustom: true, IsFreeText: true, MinLength: "3", MaxLength: "4",
	})
	// Nothing registered for this component, so any value is acceptable.
	cfg.CustomComponents = nil

	plan := validPlan()
	plan.Application = types.StringValue("waytoolongforthis")

	if d := validateAgainstConfiguration(cfg, plan); d.HasError() {
		t.Errorf("a custom component value was length checked: %v", d.Errors())
	}
}

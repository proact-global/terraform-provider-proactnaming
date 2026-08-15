// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/proact-global/azurenamingtool-client-go"
)

// Checking a request against the naming tool's own configuration before sending
// it turns failures that would otherwise arrive from the API, phrased in its
// terms and attributed to nothing in particular, into diagnostics attached to
// the argument at fault.
//
// Only values the configuration can decide are checked. A component marked free
// text accepts anything, so only its length is checked; a component that is not
// free text has an enumerable set of accepted short names, so a value can be
// rejected outright and the alternatives listed.
//
// Components the deployment has disabled are skipped: a value supplied for a
// component that takes no part in a name is not the provider's business to
// reject.

// componentNames are the naming tool's own names for the components this
// resource's arguments map onto.
const (
	componentResourceType     = "ResourceType"
	componentResourceOrg      = "ResourceOrg"
	componentResourceLocation = "ResourceLocation"
	componentResourceEnv      = "ResourceEnvironment"
	componentResourceFunction = "ResourceFunction"
	componentResourceInstance = "ResourceInstance"
)

// maxListedValues caps how many accepted values an error message quotes. A
// deployment can define hundreds of locations, and a diagnostic listing all of
// them buries the point.
const maxListedValues = 12

// validateAgainstConfiguration checks a planned resource against the naming
// tool's configuration and returns diagnostics describing anything it can
// establish is wrong.
func validateAgainstConfiguration(
	cfg *azurenamingtool.NamingConfiguration,
	plan generateNameModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// The resource type is not a component value but the selector for the whole
	// name, and the API matches it exactly and case sensitively.
	validateResourceType(cfg, plan.ResourceType, &diags)

	// Each remaining argument is checked only if its component is enabled.
	for _, check := range []struct {
		component string
		attribute string
		value     types.String
		accepted  []azurenamingtool.ComponentValue
	}{
		{componentResourceOrg, "organization", plan.Organization, cfg.Orgs},
		{componentResourceLocation, "location", plan.Location, cfg.Locations},
		{componentResourceEnv, "environment", plan.Environment, cfg.Environments},
		{componentResourceFunction, "function", plan.Function, cfg.Functions},
	} {
		validateComponentValue(cfg, check.component, check.attribute, check.value, check.accepted, &diags)
	}

	validateInstance(cfg, plan.Instance, &diags)
	validateCustomComponents(cfg, plan, &diags)

	return diags
}

func validateResourceType(cfg *azurenamingtool.NamingConfiguration, value types.String, diags *diag.Diagnostics) {
	if !known(value) {
		return
	}
	got := value.ValueString()

	var accepted []string
	for _, rt := range cfg.ResourceTypes {
		if rt.ShortName == "" {
			continue
		}
		if rt.ShortName == got {
			return // exact match, as the API requires
		}
		accepted = append(accepted, rt.ShortName)
	}

	detail := fmt.Sprintf("The Azure Naming Tool has no resource type with the short name %q.", got)

	// A case-only mismatch is worth calling out, because the API compares short
	// names exactly and the failure is otherwise baffling.
	for _, a := range accepted {
		if strings.EqualFold(a, got) {
			detail += fmt.Sprintf("\n\nIt does have %q. Short names are matched exactly, including case.", a)
			diags.AddAttributeError(path.Root("resource_type"), "Unknown Resource Type", detail)
			return
		}
	}

	detail += "\n\n" + listValues(accepted)
	detail += "\n\nThe data source proactnaming_resource_types lists every configured type."
	diags.AddAttributeError(path.Root("resource_type"), "Unknown Resource Type", detail)
}

// validateComponentValue checks one argument against the values its component
// accepts, when the component is enabled and enumerable.
func validateComponentValue(
	cfg *azurenamingtool.NamingConfiguration,
	componentName, attribute string,
	value types.String,
	accepted []azurenamingtool.ComponentValue,
	diags *diag.Diagnostics,
) {
	if !known(value) {
		return
	}
	component, ok := cfg.Component(componentName)
	if !ok || !component.Enabled {
		return
	}

	got := value.ValueString()

	// An empty optional argument is simply absent, not an invalid value.
	if got == "" {
		return
	}

	if component.IsFreeText {
		validateLength(component, attribute, got, diags)
		return
	}

	names := azurenamingtool.ShortNames(accepted)
	for _, n := range names {
		if n == got {
			return
		}
	}

	detail := fmt.Sprintf("The Azure Naming Tool does not accept %q for %s.", got, component.Name)
	for _, n := range names {
		if strings.EqualFold(n, got) {
			detail += fmt.Sprintf("\n\nIt does accept %q. Values are matched exactly, including case.", n)
			diags.AddAttributeError(path.Root(attribute), "Unknown Value for "+attribute, detail)
			return
		}
	}
	detail += "\n\n" + listValues(names)
	diags.AddAttributeError(path.Root(attribute), "Unknown Value for "+attribute, detail)
}

// validateInstance checks the instance against its component's length bounds.
// Instance is free text, so its value cannot be enumerated, but a wrong length
// is the failure people actually hit -- the tool requires an exact width and
// rejects an unpadded number.
func validateInstance(cfg *azurenamingtool.NamingConfiguration, value types.String, diags *diag.Diagnostics) {
	if !known(value) {
		return
	}
	component, ok := cfg.Component(componentResourceInstance)
	if !ok || !component.Enabled {
		return
	}
	validateLength(component, "instance", value.ValueString(), diags)
}

func validateLength(component azurenamingtool.ResourceComponent, attribute, got string, diags *diag.Diagnostics) {
	if got == "" {
		return
	}

	if component.Alphanumeric && !isAlphanumeric(got) {
		diags.AddAttributeError(
			path.Root(attribute),
			"Invalid Value for "+attribute,
			fmt.Sprintf("%s accepts letters and digits only, but %q contains other characters.",
				component.Name, got),
		)
		return
	}

	minLen, maxLen, ok := component.LengthRange()
	if !ok {
		return // the deployment sets no bounds
	}
	if len(got) >= minLen && len(got) <= maxLen {
		return
	}

	var expectation string
	if minLen == maxLen {
		expectation = fmt.Sprintf("exactly %d characters", minLen)
	} else {
		expectation = fmt.Sprintf("between %d and %d characters", minLen, maxLen)
	}

	detail := fmt.Sprintf("%s must be %s, but %q is %d.",
		component.Name, expectation, got, len(got))

	// The overwhelmingly common case: a number that has not been padded.
	if len(got) < minLen && isDigits(got) {
		detail += fmt.Sprintf("\n\nPad it, for example %q.",
			strings.Repeat("0", minLen-len(got))+got)
	}

	diags.AddAttributeError(path.Root(attribute), "Invalid Length for "+attribute, detail)
}

// validateCustomComponents checks the custom_component blocks, and the
// application argument, which the provider sends as a custom component.
func validateCustomComponents(
	cfg *azurenamingtool.NamingConfiguration,
	plan generateNameModel,
	diags *diag.Diagnostics,
) {
	// Custom component parents that the deployment defines at all.
	parents := map[string]bool{}
	for _, cc := range cfg.CustomComponents {
		parents[strings.ToLower(cc.ParentComponent)] = true
	}

	check := func(name, value string, at path.Path) {
		if name == "" || value == "" {
			return
		}
		if !parents[strings.ToLower(name)] {
			// Unknown parents are reported only when the deployment defines
			// custom components at all; otherwise the API is the better judge.
			if len(parents) == 0 {
				return
			}
			known := make([]string, 0, len(parents))
			for p := range parents {
				known = append(known, p)
			}
			sort.Strings(known)
			diags.AddAttributeError(at, "Unknown Custom Component",
				fmt.Sprintf("The Azure Naming Tool defines no custom component named %q.\n\n%s",
					name, listValues(known)))
			return
		}

		accepted := cfg.CustomComponentValues(name)
		for _, a := range accepted {
			if a == value {
				return
			}
		}
		detail := fmt.Sprintf("The custom component %q does not accept %q.", name, value)
		for _, a := range accepted {
			if strings.EqualFold(a, value) {
				detail += fmt.Sprintf("\n\nIt does accept %q. Values are matched exactly, including case.", a)
				diags.AddAttributeError(at, "Unknown Custom Component Value", detail)
				return
			}
		}
		detail += "\n\n" + listValues(accepted)
		diags.AddAttributeError(at, "Unknown Custom Component Value", detail)
	}

	if known(plan.Application) {
		check("application", plan.Application.ValueString(), path.Root("application"))
	}
	for i, c := range plan.CustomComponents {
		if !known(c.Name) || !known(c.Value) {
			continue
		}
		check(c.Name.ValueString(), c.Value.ValueString(),
			path.Root("custom_component").AtListIndex(i).AtName("value"))
	}
}

// known reports whether a value is settled enough to check. Unknown values are
// resolved later in the plan, and null ones were not supplied.
func known(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func isAlphanumeric(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// listValues renders accepted values for a diagnostic, sorted and capped so a
// deployment with hundreds of locations does not bury the message.
func listValues(values []string) string {
	if len(values) == 0 {
		return "The naming tool lists no accepted values for this component."
	}

	sorted := append([]string(nil), values...)
	sort.Strings(sorted)

	if len(sorted) <= maxListedValues {
		return "Accepted values: " + strings.Join(sorted, ", ") + "."
	}
	return fmt.Sprintf("Accepted values include: %s, and %d more.",
		strings.Join(sorted[:maxListedValues], ", "), len(sorted)-maxListedValues)
}

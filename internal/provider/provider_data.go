// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/proact-global/azurenamingtool-client-go"
)

// providerData is what the provider hands to its resources and data sources.
//
// It exists because there is now more to pass than the client alone. Passing the
// client directly would mean smuggling settings through package-level state or
// re-reading the environment in each resource, both of which hide where a
// setting came from.
type providerData struct {
	client *azurenamingtool.Client

	// predictNamesLocally selects how the name shown during a plan is arrived
	// at. See the provider schema for what the choice means.
	predictNamesLocally bool
}

// providerDataFrom recovers the provider data given to a resource or data
// source, reporting the mismatch clearly if the framework ever hands over
// something else.
func providerDataFrom(raw any) (*providerData, error) {
	if raw == nil {
		// Terraform supplies this after the ConfigureProvider RPC, so a nil here
		// is the framework's ordering rather than a fault.
		return nil, nil
	}
	data, ok := raw.(*providerData)
	if !ok {
		return nil, fmt.Errorf(
			"expected *providerData, got %T. Please report this to the provider developers", raw)
	}
	return data, nil
}

// configureResource wires provider data into a resource, reporting any problem
// through the response.
func configureResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse) (*providerData, bool) {
	data, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", err.Error())
		return nil, false
	}
	return data, data != nil
}

// configureDataSource is configureResource for data sources, which carry their
// own request and response types.
func configureDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) (*providerData, bool) {
	data, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", err.Error())
		return nil, false
	}
	return data, data != nil
}

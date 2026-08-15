// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/proact-global/azurenamingtool-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &azurermResourcesDataSource{}
	_ datasource.DataSourceWithConfigure = &azurermResourcesDataSource{}
)

// NewAzurermResourcesDataSource is a helper function to simplify the provider implementation.
func NewAzurermResourcesDataSource() datasource.DataSource {
	return &azurermResourcesDataSource{}
}

// azurermResourcesDataSource is the data source implementation.
type azurermResourcesDataSource struct {
	client *azurenamingtool.Client
}

// azurermResourcesDataSourceModel maps the data source schema data.
type azurermResourcesDataSourceModel struct {
	Resources types.Map `tfsdk:"resources"`
}

// Metadata returns the data source type name.
func (d *azurermResourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azurerm_resources"
}

// Schema defines the schema for the data source.
func (d *azurermResourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns a map of azurerm Terraform resource types to their short names.",
		MarkdownDescription: "Returns a map of azurerm Terraform resource types to their short names. " +
			"Use this data source to easily look up the short name for a given azurerm resource type " +
			"(e.g. `azurerm_storage_account` => `st`). Only resource types without a property qualifier are included.",
		Attributes: map[string]schema.Attribute{
			"resources": schema.MapAttribute{
				Description: "Map of azurerm resource type (e.g. `azurerm_storage_account`) to its short name (e.g. `st`).",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *azurermResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state azurermResourcesDataSourceModel

	resourceTypes, err := d.client.GetResourceTypes()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read azurenamingtool resource_types",
			err.Error(),
		)
		return
	}

	mapping := make(map[string]string)
	for _, rt := range resourceTypes {
		if rt.Property != "" {
			continue
		}
		azurermType := getAzurermResourceType(rt.Resource, rt.Property)
		if azurermType == "" {
			continue
		}
		for _, part := range strings.Split(azurermType, ",") {
			key := strings.TrimSpace(part)
			if key != "" {
				mapping[key] = rt.ShortName
			}
		}
	}

	resourcesMap, diags := types.MapValueFrom(ctx, types.StringType, mapping)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Resources = resourcesMap

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *azurermResourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data, ok := configureDataSource(req, resp)
	if !ok {
		return
	}

	d.client = data.client
}

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
  "github.com/proact-global/azurenamingtool-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &resourceTypesDataSource{}
	_ datasource.DataSourceWithConfigure = &resourceTypesDataSource{}
)

// NewresourceTypesDataSource is a helper function to simplify the provider implementation.
func NewresourceTypesDataSource() datasource.DataSource {
	return &resourceTypesDataSource{}
}

// resourceTypesDataSource is the data source implementation.
type resourceTypesDataSource struct {
	client *azurenamingtool.Client
}

// resourceTypesDataSourceModel maps the data source schema data.
type resourceTypesDataSourceModel struct {
	ResourceTypes []resourceTypesModel `tfsdk:"resource_types"`
}

// resourceTypesModel maps resourceTypes schema data.
type resourceTypesModel struct {
	ID                           types.Int64  `tfsdk:"id"`
	Resource                     types.String `tfsdk:"resource"`
	Optional                     types.String `tfsdk:"optional"`
	Exclude                      types.String `tfsdk:"exclude"`
	Property                     types.String `tfsdk:"property"`
	ShortName                    types.String `tfsdk:"short_name"`
	Scope                        types.String `tfsdk:"scope"`
	LengthMin                    types.String `tfsdk:"length_min"`
	LengthMax                    types.String `tfsdk:"length_max"`
	ValidText                    types.String `tfsdk:"valid_text"`
	InvalidText                  types.String `tfsdk:"invalid_text"`
	InvalidCharacters            types.String `tfsdk:"invalid_characters"`
	InvalidCharactersStart       types.String `tfsdk:"invalid_characters_start"`
	InvalidCharactersEnd         types.String `tfsdk:"invalid_characters_end"`
	InvalidCharactersConsecutive types.String `tfsdk:"invalid_characters_consecutive"`
	Regx                         types.String `tfsdk:"regx"`
	StaticValues                 types.String `tfsdk:"static_values"`
	Enabled                      bool         `tfsdk:"enabled"`
	ApplyDelimiter               bool         `tfsdk:"apply_delimiter"`
}

// Metadata returns the data source type name.
func (d *resourceTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_types"
}

// Schema defines the schema for the data source.
func (d *resourceTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_types": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"resource": schema.StringAttribute{
							Computed: true,
						},
						"optional": schema.StringAttribute{
							Computed: true,
						},
						"exclude": schema.StringAttribute{
							Computed: true,
						},
						"property": schema.StringAttribute{
							Computed: true,
						},
						"short_name": schema.StringAttribute{
							Computed: true,
						},
						"scope": schema.StringAttribute{
							Computed: true,
						},
						"length_min": schema.StringAttribute{
							Computed: true,
						},
						"length_max": schema.StringAttribute{
							Computed: true,
						},
						"valid_text": schema.StringAttribute{
							Computed: true,
						},
						"invalid_text": schema.StringAttribute{
							Computed: true,
						},
						"invalid_characters": schema.StringAttribute{
							Computed: true,
						},
						"invalid_characters_start": schema.StringAttribute{
							Computed: true,
						},
						"invalid_characters_end": schema.StringAttribute{
							Computed: true,
						},
						"invalid_characters_consecutive": schema.StringAttribute{
							Computed: true,
						},
						"regx": schema.StringAttribute{
							Computed: true,
						},
						"static_values": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"apply_delimiter": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *resourceTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state resourceTypesDataSourceModel

	resourceTypes, err := d.client.GetResourceTypes()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read azurenamingtool resource_types",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, resource_types := range resourceTypes {
		resourceTypestate := resourceTypesModel{
			ID:                           types.Int64Value(int64(resource_types.ID)),
			Resource:                     types.StringValue(resource_types.Resource),
			Optional:                     types.StringValue(resource_types.Optional),
			Exclude:                      types.StringValue(resource_types.Exclude),
			Property:                     types.StringValue(resource_types.Property),
			ShortName:                    types.StringValue(resource_types.ShortName),
			Scope:                        types.StringValue(resource_types.Scope),
			LengthMin:                    types.StringValue(resource_types.LengthMin),
			LengthMax:                    types.StringValue(resource_types.LengthMax),
			ValidText:                    types.StringValue(resource_types.ValidText),
			InvalidText:                  types.StringValue(resource_types.InvalidText),
			InvalidCharacters:            types.StringValue(resource_types.InvalidCharacters),
			InvalidCharactersStart:       types.StringValue(resource_types.InvalidCharactersStart),
			InvalidCharactersEnd:         types.StringValue(resource_types.InvalidCharactersEnd),
			InvalidCharactersConsecutive: types.StringValue(resource_types.InvalidCharactersConsecutive),
			Regx:                         types.StringValue(resource_types.Regx),
			StaticValues:                 types.StringValue(resource_types.StaticValues),
			Enabled:                      resource_types.Enabled,
			ApplyDelimiter:               resource_types.ApplyDelimiter,
		}
		state.ResourceTypes = append(state.ResourceTypes, resourceTypestate)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *resourceTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*azurenamingtool.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *azurenamingtool.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

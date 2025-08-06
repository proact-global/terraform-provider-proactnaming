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
	_ datasource.DataSource              = &GeneratedNameDataSource{}
	_ datasource.DataSourceWithConfigure = &GeneratedNameDataSource{}
)

// NewGeneratedNameDataSource is a helper function to simplify the provider implementation.
func NewGeneratedNameDataSource() datasource.DataSource {
	return &GeneratedNameDataSource{}
}

// GeneratedNameDataSource is the data source implementation.
type GeneratedNameDataSource struct {
	client *azurenamingtool.Client
}

// GeneratedNameDataSourceModel maps the data source schema data.
type GeneratedNameDataSourceModel struct {
	ID            types.Int64          `tfsdk:"id"`
	GeneratedName []GeneratedNameModel `tfsdk:"generated_name"`
}

// GeneratedNameModel maps GeneratedName schema data.
type GeneratedNameModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ResourceName     types.String `tfsdk:"resource_name"`
	ResourceTypeName types.String `tfsdk:"resource_type_name"`
}

// Metadata returns the data source type name.
func (d *GeneratedNameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_generated_name"
}

// Schema defines the schema for the data source.
func (d *GeneratedNameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the generated name data source.",
				Required:    true,
			},
			"generated_name": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"resource_name": schema.StringAttribute{
							Computed: true,
						},
						"resource_type_name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *GeneratedNameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GeneratedNameDataSourceModel

	// Read config to get the ID
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueInt64()
	id16 := int16(id)

	generatedName, err := d.client.GetName(id16)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read azurenamingtool generated_name",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.GeneratedName = nil // reset before appending
	if generatedName != nil {
		generatedNameState := GeneratedNameModel{
			ID:               types.Int64Value(int64(generatedName.ID)),
			ResourceName:     types.StringValue(generatedName.ResourceName),
			ResourceTypeName: types.StringValue(generatedName.ResourceTypeName),
		}
		state.GeneratedName = append(state.GeneratedName, generatedNameState)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *GeneratedNameDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/proact-global/azurenamingtool-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &generateName{}
	_ resource.ResourceWithConfigure = &generateName{}
)

// NewGenerateName is a helper function to simplify the provider implementation.
func NewGenerateName() resource.Resource {
	return &generateName{}
}

// generateName is the resource implementation.
type generateName struct {
	client *azurenamingtool.Client
}

// generateNameModel maps the resource schema data.
type generateNameModel struct {
	// Input fields for name generation
	Organization types.String `tfsdk:"organization"`
	ResourceType types.String `tfsdk:"resource_type"`
	Application  types.String `tfsdk:"application"`
	Function     types.String `tfsdk:"function"`
	Instance     types.String `tfsdk:"instance"`
	Location     types.String `tfsdk:"location"`
	Environment  types.String `tfsdk:"environment"`

	// Output fields from the API
	ID           types.Int64  `tfsdk:"id"`
	ResourceName types.String `tfsdk:"resource_name"`
	Success      types.Bool   `tfsdk:"success"`
	Message      types.String `tfsdk:"message"`
}

// Metadata returns the resource type name.
func (r *generateName) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_generate_name"
}

// Schema defines the schema for the resource.
func (r *generateName) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Generates standardized Azure resource names using the Azure Naming Tool.",
		MarkdownDescription: "Generates standardized Azure resource names using the Azure Naming Tool following organizational naming conventions.",
		Attributes: map[string]schema.Attribute{
			// Input attributes
			"organization": schema.StringAttribute{
				Description: "Organization identifier for the resource name.",
				Required:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "Azure resource type short name (e.g., 'rg', 'st', 'vm').",
				Required:    true,
			},
			"application": schema.StringAttribute{
				Description: "Application identifier for the resource name.",
				Required:    true,
			},
			"function": schema.StringAttribute{
				Description: "Function or purpose identifier for the resource name.",
				Optional:    true,
			},
			"instance": schema.StringAttribute{
				Description: "Instance number or identifier for the resource name.",
				Required:    true,
			},
			"location": schema.StringAttribute{
				Description: "Azure region identifier (e.g., 'euw', 'eus').",
				Required:    true,
			},
			"environment": schema.StringAttribute{
				Description: "Environment identifier (e.g., 'dev', 'test', 'prod').",
				Required:    true,
			},

			// Output attributes
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the generated name in the Azure Naming Tool.",
				Computed:    true,
			},
			"resource_name": schema.StringAttribute{
				Description: "The generated Azure resource name.",
				Computed:    true,
			},
			"success": schema.BoolAttribute{
				Description: "Indicates whether the name generation was successful.",
				Computed:    true,
			},
			"message": schema.StringAttribute{
				Description: "Message from the Azure Naming Tool API.",
				Computed:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *generateName) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan generateNameModel

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if we already have an ID from Read (planning phase)
	if !plan.ID.IsNull() && !plan.ID.IsUnknown() {
		// Use the existing ID from the planning phase
		id := plan.ID.ValueInt64()
		id16 := int16(id)

		// Get the name details using the ID
		nameDetails, err := r.client.GetName(id16)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Retrieve Generated Name",
				fmt.Sprintf("An error occurred while retrieving the generated name with ID %d: %s", id, err.Error()),
			)
			return
		}

		// Set the final state
		plan.ID = types.Int64Value(int64(nameDetails.ID))
		plan.ResourceName = types.StringValue(nameDetails.ResourceName)
		plan.Success = types.BoolValue(true)
		plan.Message = types.StringValue("Name successfully created and retrieved")
	} else {
		// Fallback: Generate the name if no ID exists (shouldn't happen in normal flow)
		generateRequest := azurenamingtool.GenerateNameRequest{
			ResourceOrg:         plan.Organization.ValueString(),
			ResourceType:        plan.ResourceType.ValueString(),
			ResourceEnvironment: plan.Environment.ValueString(),
			ResourceFunction:    plan.Function.ValueString(),
			ResourceInstance:    plan.Instance.ValueString(),
			ResourceLocation:    plan.Location.ValueString(),
			CustomComponents: azurenamingtool.GenerateNameRequestCustomComponents{
				Application: plan.Application.ValueString(),
			},
		}

		generateResponse, err := r.client.GenerateName(generateRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Generate Name",
				fmt.Sprintf("An error occurred while generating the name: %s", err.Error()),
			)
			return
		}

		// Set the generated values
		plan.ID = types.Int64Value(generateResponse.ResourceNameDetails.ID)
		plan.ResourceName = types.StringValue(generateResponse.ResourceName)
		plan.Success = types.BoolValue(generateResponse.Success)
		plan.Message = types.StringValue(generateResponse.Message)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
// This function generates the name during planning to show the result in terraform plan
func (r *generateName) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state generateNameModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If we already have an ID, use GetName to get the current state
	if !state.ID.IsNull() && !state.ID.IsUnknown() {
		id := state.ID.ValueInt64()
		id16 := int16(id)

		nameDetails, err := r.client.GetName(id16)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Generated Name",
				fmt.Sprintf("An error occurred while reading the generated name with ID %d: %s", id, err.Error()),
			)
			return
		}

		// Update state with current data
		state.ID = types.Int64Value(int64(nameDetails.ID))
		state.ResourceName = types.StringValue(nameDetails.ResourceName)
		state.Success = types.BoolValue(true)
		state.Message = types.StringValue("Name successfully retrieved")
	} else {
		// This is for planning - generate the name to show what will be created
		generateRequest := azurenamingtool.GenerateNameRequest{
			ResourceOrg:         state.Organization.ValueString(),
			ResourceType:        state.ResourceType.ValueString(),
			ResourceEnvironment: state.Environment.ValueString(),
			ResourceFunction:    state.Function.ValueString(),
			ResourceInstance:    state.Instance.ValueString(),
			ResourceLocation:    state.Location.ValueString(),
			CustomComponents: azurenamingtool.GenerateNameRequestCustomComponents{
				Application: state.Application.ValueString(),
			},
		}

		generateResponse, err := r.client.GenerateName(generateRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Generate Name",
				fmt.Sprintf("An error occurred while generating the name: %s", err.Error()),
			)
			return
		}

		// Store the generated name details for use in Create
		state.ID = types.Int64Value(generateResponse.ResourceNameDetails.ID)
		state.ResourceName = types.StringValue(generateResponse.ResourceName)
		state.Success = types.BoolValue(generateResponse.Success)
		state.Message = types.StringValue(generateResponse.Message)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *generateName) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan generateNameModel

	// Retrieve values from plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate a new name with the updated parameters
	generateRequest := azurenamingtool.GenerateNameRequest{
		ResourceOrg:         plan.Organization.ValueString(),
		ResourceType:        plan.ResourceType.ValueString(),
		ResourceEnvironment: plan.Environment.ValueString(),
		ResourceFunction:    plan.Function.ValueString(),
		ResourceInstance:    plan.Instance.ValueString(),
		ResourceLocation:    plan.Location.ValueString(),
		CustomComponents: azurenamingtool.GenerateNameRequestCustomComponents{
			Application: plan.Application.ValueString(),
		},
	}

	generateResponse, err := r.client.GenerateName(generateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Generate Updated Name",
			fmt.Sprintf("An error occurred while generating the updated name: %s", err.Error()),
		)
		return
	}

	// Update the state with new values
	plan.ID = types.Int64Value(generateResponse.ResourceNameDetails.ID)
	plan.ResourceName = types.StringValue(generateResponse.ResourceName)
	plan.Success = types.BoolValue(generateResponse.Success)
	plan.Message = types.StringValue(generateResponse.Message)

	// Set updated state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *generateName) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The Azure Naming Tool API doesn't support deletion via API key authentication.
	// Delete operations require both an API key and a plain text password, which is not secure.
	// For now, we'll just remove from Terraform state without calling the API.

	// Add a warning to let users know about this limitation
	resp.Diagnostics.AddWarning(
		"Delete Operation Limitation",
		"The Azure Naming Tool API doesn't support deletion via API key authentication. "+
			"The generated name remains in the naming tool but has been removed from Terraform state. "+
			"If you need to delete the name from the naming tool, you'll need to do it manually through the web interface.",
	)

	// The resource is automatically removed from state when this function completes successfully
}

// Configure adds the provider configured client to the resource.
func (r *generateName) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*azurenamingtool.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *azurenamingtool.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

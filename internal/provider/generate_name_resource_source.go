package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/proact-global/azurenamingtool-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource               = &generateName{}
	_ resource.ResourceWithConfigure  = &generateName{}
	_ resource.ResourceWithModifyPlan = &generateName{}
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
		Description: "Generates standardized Azure resource names using the Azure Naming Tool.",
		MarkdownDescription: "Generates standardized Azure resource names using the Azure Naming Tool following organizational naming conventions.\n\n" +
			"This resource creates names that comply with Azure naming rules and organizational standards. " +
			"All input fields trigger resource replacement when changed, ensuring name consistency.",
		Attributes: map[string]schema.Attribute{
			// Input attributes
			"organization": schema.StringAttribute{
				Description:   "Organization identifier for the resource name.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"resource_type": schema.StringAttribute{
				Description:   "Azure resource type short name (e.g., 'rg', 'st', 'vm').",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"application": schema.StringAttribute{
				Description:   "Application identifier for the resource name.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"function": schema.StringAttribute{
				Description:   "Function or purpose identifier for the resource name.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"instance": schema.StringAttribute{
				Description:   "Instance number or identifier for the resource name.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"location": schema.StringAttribute{
				Description:   "Azure region identifier (e.g., 'euw', 'eus').",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"environment": schema.StringAttribute{
				Description:   "Environment identifier (e.g., 'dev', 'test', 'prod').",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
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

	// Generate the name using the API
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
			fmt.Sprintf("An error occurred while generating the name: %s\n\n"+
				"Please verify:\n"+
				"- Azure Naming Tool is accessible at the configured host\n"+
				"- API key has sufficient permissions\n"+
				"- Input parameters match your naming tool configuration", err.Error()),
		)
		return
	}

	// Set the generated values in state
	plan.ID = types.Int64Value(generateResponse.ResourceNameDetails.ID)
	plan.ResourceName = types.StringValue(generateResponse.ResourceName)
	plan.Success = types.BoolValue(generateResponse.Success)
	plan.Message = types.StringValue(generateResponse.Message)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *generateName) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state generateNameModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For this resource, we don't need to fetch from a remote API since the name
	// is generated once and stored in state. The Azure Naming Tool doesn't maintain
	// persistent storage of generated names for retrieval.

	// We keep the generated entry in the Azure Naming Tool until explicit destroy
	// This ensures we maintain the entry for the lifetime of the Terraform resource
	// Only preview entries from ModifyPlan are cleaned up immediately
} // Update updates the resource and sets the updated Terraform state on success.
// Note: This function should not be called due to RequiresReplace plan modifiers
// on all input attributes, but is implemented for completeness.
func (r *generateName) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since all input attributes have RequiresReplace plan modifiers,
	// this function should not be called. If it is called, we'll return an error.
	resp.Diagnostics.AddError(
		"Unexpected Update Operation",
		"This resource does not support updates. All changes should trigger a replacement. "+
			"If you see this error, please report it to the provider developers.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *generateName) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state generateNameModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is null, it means the entry was already cleaned up during Read
	// This is expected behavior for our cleanup strategy, so we just return success
	if state.ID.IsNull() || state.ID.IsUnknown() {
		// Resource was already cleaned up, nothing to delete
		return
	}

	// Delete the generated name using the ID
	id := state.ID.ValueInt64()

	deleteRequest := azurenamingtool.DeleteGeneratedNameRequest{
		ID: id,
	}

	_, err := r.client.DeleteName(deleteRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Generated Name",
			fmt.Sprintf("An error occurred while deleting the generated name with ID %d: %s\n\n"+
				"This may indicate:\n"+
				"- The entry was already deleted\n"+
				"- Admin password is required but not configured\n"+
				"- Network connectivity issues with the Azure Naming Tool", id, err.Error()),
		)
		return
	}

	// The resource is automatically removed from state when this function completes successfully
}

// ModifyPlan generates the name during planning to show what the new name will be
func (r *generateName) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip for destroy operations
	if req.Plan.Raw.IsNull() {
		return
	}

	// Only generate names for create operations (when state is null) or updates
	var plan generateNameModel

	// Get the planned configuration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip if we don't have all required fields
	if plan.Organization.IsNull() || plan.ResourceType.IsNull() ||
		plan.Application.IsNull() || plan.Instance.IsNull() ||
		plan.Location.IsNull() || plan.Environment.IsNull() {
		return
	}

	// Skip if the client is not available (shouldn't happen, but safety check)
	if r.client == nil {
		return
	}

	// Only update unknown computed values
	shouldGenerate := false
	if req.State.Raw.IsNull() {
		// This is a create operation - generate if resource_name is unknown
		shouldGenerate = plan.ResourceName.IsUnknown()
	} else {
		// This is an update/replace operation - only generate if there are changes to inputs
		var state generateNameModel
		diags := req.State.Get(ctx, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Check if any input has changed (which would trigger RequiresReplace)
		inputChanged := !plan.Organization.Equal(state.Organization) ||
			!plan.ResourceType.Equal(state.ResourceType) ||
			!plan.Application.Equal(state.Application) ||
			!plan.Function.Equal(state.Function) ||
			!plan.Instance.Equal(state.Instance) ||
			!plan.Location.Equal(state.Location) ||
			!plan.Environment.Equal(state.Environment)

		shouldGenerate = inputChanged && plan.ResourceName.IsUnknown()
	}

	if !shouldGenerate {
		return
	}

	// Generate the name using the API
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
		// Don't fail the plan, just skip showing the generated name
		return
	}

	// Update the plan with the generated values to show in the plan output
	plan.ResourceName = types.StringValue(generateResponse.ResourceName)
	// Note: We don't set ID here since it might be different in the actual create

	// Clean up the preview entry immediately to prevent accumulation in the tool
	// This is a "preview" entry that shouldn't persist
	if generateResponse.ResourceNameDetails.ID != 0 {
		deleteRequest := azurenamingtool.DeleteGeneratedNameRequest{
			ID: generateResponse.ResourceNameDetails.ID,
		}
		// Ignore errors - this is cleanup for preview entries
		_, _ = r.client.DeleteName(deleteRequest)
	}

	// Set the updated plan
	diags = resp.Plan.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
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

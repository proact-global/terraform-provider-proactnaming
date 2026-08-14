// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"errors"
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

// customComponentModel maps a single custom_component block.
type customComponentModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// generateNameModel maps the resource schema data.
type generateNameModel struct {
	// Input fields for name generation.
	Organization     types.String           `tfsdk:"organization"`
	ResourceType     types.String           `tfsdk:"resource_type"`
	Application      types.String           `tfsdk:"application"`
	Function         types.String           `tfsdk:"function"`
	Instance         types.String           `tfsdk:"instance"`
	Location         types.String           `tfsdk:"location"`
	Environment      types.String           `tfsdk:"environment"`
	CustomComponents []customComponentModel `tfsdk:"custom_component"`

	// Output fields from the API.
	ID           types.Int64  `tfsdk:"id"`
	ResourceName types.String `tfsdk:"resource_name"`
	Success      types.Bool   `tfsdk:"success"`
	Message      types.String `tfsdk:"message"`
}

// buildCustomComponents converts the model's custom_component blocks and the
// application field into the map expected by the API client.
func (m *generateNameModel) buildCustomComponents() azurenamingtool.GenerateNameRequestCustomComponents {
	cc := make(azurenamingtool.GenerateNameRequestCustomComponents)
	if !m.Application.IsNull() && !m.Application.IsUnknown() {
		cc["application"] = m.Application.ValueString()
	}
	for _, c := range m.CustomComponents {
		if !c.Name.IsNull() && !c.Name.IsUnknown() && !c.Value.IsNull() && !c.Value.IsUnknown() {
			cc[c.Name.ValueString()] = c.Value.ValueString()
		}
	}
	return cc
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
			// Input attributes.
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

			// Output attributes.
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
		Blocks: map[string]schema.Block{
			"custom_component": schema.ListNestedBlock{
				Description: "Additional custom naming components required by certain resource types (e.g. subnet_tier, subnet_instance for subnets). " +
					"Each block specifies a component name and its short-name value. " +
					"Use multiple blocks — or a `dynamic` block — when a resource requires more than one custom component.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:   "Custom component name as defined in the Azure Naming Tool (e.g. 'subnet_tier', 'subnet_instance').",
							Required:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"value": schema.StringAttribute{
							Description:   "Short-name value for the custom component (e.g. 'app', 'web', '01').",
							Required:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *generateName) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan generateNameModel

	// Retrieve values from plan.
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Now we actually generate and persist the name during Create (apply phase).
	// This creates the persistent entry in Azure Naming Tool.
	generateRequest := azurenamingtool.GenerateNameRequest{
		ResourceOrg:         plan.Organization.ValueString(),
		ResourceType:        plan.ResourceType.ValueString(),
		ResourceEnvironment: plan.Environment.ValueString(),
		ResourceFunction:    plan.Function.ValueString(),
		ResourceInstance:    plan.Instance.ValueString(),
		ResourceLocation:    plan.Location.ValueString(),
		CustomComponents:    plan.buildCustomComponents(),
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

	// Set the generated values in state - this creates the persistent entry.
	plan.ID = types.Int64Value(generateResponse.ResourceNameDetails.ID)
	plan.ResourceName = types.StringValue(generateResponse.ResourceName)
	plan.Success = types.BoolValue(generateResponse.Success)
	plan.Message = types.StringValue(generateResponse.Message)

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *generateName) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state generateNameModel

	// Get current state.
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If we have an ID, verify that the name still exists in the naming tool.
	// This handles the case where a name was deleted outside of Terraform.
	if !state.ID.IsNull() && !state.ID.IsUnknown() {
		id := state.ID.ValueInt64()
		nameDetails, err := r.client.GetName(id)
		if err != nil {
			if errors.Is(err, azurenamingtool.ErrNotFound) {
				// Name no longer exists in the naming tool; remove from state so
				// Terraform knows to recreate it on the next apply.
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError(
				"Unable to Read Generated Name",
				fmt.Sprintf("An error occurred while reading the generated name with ID %d: %s", id, err.Error()),
			)
			return
		}
		// Update state with current values from the naming tool.
		state.ResourceName = types.StringValue(nameDetails.ResourceName)
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		return
	}

	// No ID in state yet — generate the name to populate missing fields.
	if state.ID.IsNull() || state.ID.IsUnknown() {
		// Generate the name using the API to create a persistent entry.
		generateRequest := azurenamingtool.GenerateNameRequest{
			ResourceOrg:         state.Organization.ValueString(),
			ResourceType:        state.ResourceType.ValueString(),
			ResourceEnvironment: state.Environment.ValueString(),
			ResourceFunction:    state.Function.ValueString(),
			ResourceInstance:    state.Instance.ValueString(),
			ResourceLocation:    state.Location.ValueString(),
			CustomComponents:    state.buildCustomComponents(),
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

		// Update the state with the generated values.
		state.ID = types.Int64Value(generateResponse.ResourceNameDetails.ID)
		state.ResourceName = types.StringValue(generateResponse.ResourceName)
		state.Success = types.BoolValue(generateResponse.Success)
		state.Message = types.StringValue(generateResponse.Message)

		// Save the updated state.
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// If ID is already populated, we keep the existing generated name.
	// This ensures the name remains stable across reads unless explicitly regenerated.
}

// Update updates the resource and sets the updated Terraform state on success.
// Note: This function should not be called due to RequiresReplace plan modifiers.
// on all input attributes, but is implemented for completeness.
func (r *generateName) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since all input attributes have RequiresReplace plan modifiers.
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

	// Get current state.
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is null, it means the entry was already cleaned up during Read.
	// This is expected behavior for our cleanup strategy, so we just return success.
	if state.ID.IsNull() || state.ID.IsUnknown() {
		// Resource was already cleaned up, nothing to delete.
		return
	}

	// Delete the generated name using the ID.
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

	// The resource is automatically removed from state when this function completes successfully.
}

// ModifyPlan generates a preview of the name during planning so users can see the actual
// name before applying. It calls GenerateName then immediately deletes the preview entry.
func (r *generateName) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip for destroy operations.
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan generateNameModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only generate a preview for new/unknown names.
	if !plan.ResourceName.IsUnknown() {
		return
	}

	// Skip if required fields are missing.
	if plan.Organization.IsNull() || plan.ResourceType.IsNull() ||
		plan.Application.IsNull() || plan.Instance.IsNull() ||
		plan.Location.IsNull() || plan.Environment.IsNull() {
		return
	}

	if r.client == nil {
		return
	}

	generateRequest := azurenamingtool.GenerateNameRequest{
		ResourceOrg:         plan.Organization.ValueString(),
		ResourceType:        plan.ResourceType.ValueString(),
		ResourceEnvironment: plan.Environment.ValueString(),
		ResourceFunction:    plan.Function.ValueString(),
		ResourceInstance:    plan.Instance.ValueString(),
		ResourceLocation:    plan.Location.ValueString(),
		CustomComponents:    plan.buildCustomComponents(),
	}

	generateResponse, err := r.client.GenerateName(generateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Generate Name Preview",
			fmt.Sprintf("An error occurred while generating the name preview: %s", err.Error()),
		)
		return
	}

	// Immediately delete the preview entry — it only existed to show the name in the plan.
	//
	// A failure here does not change the name that will be applied: the naming tool is
	// configured to allow duplicate names, so a leftover entry does not cause the apply
	// to generate a different name. It does leave an orphan record behind though, and
	// those accumulate on every plan, so report it instead of discarding the error.
	//
	// This is also the only place a bad admin password becomes visible: the naming tool
	// answers an incorrect or missing password with HTTP 200 and a "FAILURE" body rather
	// than an authentication error, so nothing surfaces it at provider configuration time.
	if id := generateResponse.ResourceNameDetails.ID; id != 0 {
		if _, err := r.client.DeleteName(azurenamingtool.DeleteGeneratedNameRequest{ID: id}); err != nil {
			resp.Diagnostics.AddWarning(
				"Preview Name Cleanup Failed",
				fmt.Sprintf("The name preview generated during planning could not be removed from the "+
					"Azure Naming Tool, leaving entry ID %d behind.\n\n"+
					"The planned name is still correct and the apply can proceed, but this entry will "+
					"remain in the naming tool until it is removed manually.\n\n"+
					"The most common cause is an incorrect admin_password, which the naming tool reports "+
					"here rather than when the provider is configured.\n\n"+
					"Error: %s", id, err),
			)
		}
	}

	plan.ResourceName = types.StringValue(generateResponse.ResourceName)

	diags = resp.Plan.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the resource.
func (r *generateName) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform.
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

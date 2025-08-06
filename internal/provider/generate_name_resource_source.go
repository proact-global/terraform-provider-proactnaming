package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

// generatedName is the resource implementation.
type generateName struct {
	client *azurenamingtool.Client
}

// generateNameModel maps the resource schema data.
type generateNameModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	ResourceName        types.String `tfsdk:"resource_name"`
	ResourceTypeName    types.String `tfsdk:"resource_type_name"`
	Message             types.String `tfsdk:"message"`
	Environment         types.String `tfsdk:"environment"`
	Instance            types.String `tfsdk:"instance"`
	Location            types.String `tfsdk:"location"`
	Organization        types.String `tfsdk:"organization"`
	ResourceType        types.String `tfsdk:"resource_type"`
	Application         types.String `tfsdk:"application"`
	ResourceFunction    types.String `tfsdk:"function"`
	Success             types.Bool   `tfsdk:"success"`
	ResourceNameDetails types.Map    `tfsdk:"resource_name_details"`
}

// Metadata returns the resource type name.
func (r *generateName) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_generate_name"
}

// Schema defines the schema for the resource.
func (r *generateName) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates standardized Azure resource names using the Azure Naming Tool API. " +
			"This resource creates names following organizational naming conventions and ensures consistency across infrastructure.",
		MarkdownDescription: "Generates standardized Azure resource names using the Azure Naming Tool API.\n\n" +
			"This resource creates names following organizational naming conventions and ensures consistency across infrastructure. " +
			"The generated names comply with Azure resource naming requirements and organizational standards.\n\n" +
			"**Note:** All input parameters force replacement when changed, as generated names are immutable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier for the generated name resource.",
				Computed:    true,
			},
			"resource_name": schema.StringAttribute{
				Description: "The generated resource name following Azure naming conventions.",
				Computed:    true,
			},
			"message": schema.StringAttribute{
				Description: "Response message from the naming tool API.",
				Computed:    true,
			},
			"success": schema.BoolAttribute{
				Description: "Indicates whether the name generation was successful.",
				Computed:    true,
			},
			"resource_name_details": schema.MapAttribute{
				Description: "Detailed breakdown of the generated resource name components.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"environment": schema.StringAttribute{
				Description: "Environment designation (e.g., 'dev', 'test', 'prod').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance": schema.StringAttribute{
				Description: "Instance number for the resource (e.g., '001', '002').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"location": schema.StringAttribute{
				Description: "Azure region/location code (e.g., 'euw' for West Europe).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization": schema.StringAttribute{
				Description: "Organization identifier (e.g., 'acme', 'contoso').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type": schema.StringAttribute{
				Description: "Azure resource type abbreviation (e.g., 'st' for Storage Account, 'vm' for Virtual Machine).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type_name": schema.StringAttribute{
				Description: "Full name of the Azure resource type.",
				Computed:    true,
			},
			"application": schema.StringAttribute{
				Description: "Application name or identifier (e.g., 'web', 'api', 'db').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"function": schema.StringAttribute{
				Description: "Function or role designation (e.g., 'web', 'api'). Can be empty.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *generateName) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Starting create operation for generate_name resource")

	// Retrieve values from plan
	var plan generateNameModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to retrieve plan data", map[string]interface{}{
			"diagnostics": resp.Diagnostics,
		})
		return
	}

	tflog.Info(ctx, "Creating generate_name resource", map[string]interface{}{
		"organization":  plan.Organization.ValueString(),
		"resource_type": plan.ResourceType.ValueString(),
		"application":   plan.Application.ValueString(),
		"environment":   plan.Environment.ValueString(),
		"location":      plan.Location.ValueString(),
		"instance":      plan.Instance.ValueString(),
	})

	var items []azurenamingtool.GenerateNameRequest
	items = append(items, azurenamingtool.GenerateNameRequest{
		ResourceEnvironment: plan.Environment.ValueString(),
		ResourceFunction:    plan.ResourceFunction.ValueString(),
		ResourceInstance:    plan.Instance.ValueString(),
		ResourceLocation:    plan.Location.ValueString(),
		ResourceOrg:         plan.Organization.ValueString(),
		ResourceType:        plan.ResourceType.ValueString(),
		CustomComponents: azurenamingtool.GenerateNameRequestCustomComponents{
			Application: plan.Application.ValueString(),
		},
	})

	// Call the client to generate the name
	generateNames, err := r.client.GenerateName(items[0])
	if err != nil {
		tflog.Error(ctx, "API call to generate name failed", map[string]interface{}{
			"error": err.Error(),
		})

		// Enhanced error handling with specific error types
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "401") {
			resp.Diagnostics.AddError(
				"Authentication Failed",
				"Invalid API key or insufficient permissions. Please check your API key configuration and ensure it has the necessary permissions to generate names.",
			)
		} else if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "403") {
			resp.Diagnostics.AddError(
				"Access Forbidden",
				"The API key does not have permission to perform this operation. Please contact your administrator.",
			)
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Resource Type Not Found",
				fmt.Sprintf("The specified resource type '%s' was not found in the naming tool configuration.", plan.ResourceType.ValueString()),
			)
		} else if strings.Contains(err.Error(), "timeout") {
			resp.Diagnostics.AddError(
				"Request Timeout",
				"The request to the naming tool API timed out. Please try again or check the API endpoint status.",
			)
		} else if strings.Contains(err.Error(), "connection") {
			resp.Diagnostics.AddError(
				"Connection Error",
				"Unable to connect to the naming tool API. Please check the host configuration and network connectivity.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Create azurenamingtool generated_name",
				fmt.Sprintf("An unexpected error occurred while generating the name: %s", err.Error()),
			)
		}
		return
	}

	tflog.Debug(ctx, "Successfully received response from naming tool API", map[string]interface{}{
		"resource_name": generateNames.ResourceName,
		"success":       generateNames.Success,
	})

	detailsMap := map[string]attr.Value{}
	// detailsMap := map[string]types.Value{}
	// Check if ResourceNameDetails is not its zero value by checking a key field
	if generateNames.ResourceNameDetails.ResourceTypeName != "" {
		detailsMap["resource_type_name"] = types.StringValue(generateNames.ResourceNameDetails.ResourceTypeName)
		// Add more fields as needed from ResourceNameDetails
	}

	state := generateNameModel{
		ID:                  types.Int64Value(int64(generateNames.ResourceNameDetails.ID)),
		ResourceName:        types.StringValue(generateNames.ResourceName),
		ResourceTypeName:    types.StringValue(generateNames.ResourceNameDetails.ResourceTypeName),
		Message:             types.StringValue(generateNames.Message),
		Environment:         plan.Environment,
		Instance:            plan.Instance,
		Location:            plan.Location,
		Organization:        plan.Organization,
		ResourceType:        plan.ResourceType,
		Application:         plan.Application,
		ResourceFunction:    plan.ResourceFunction,
		Success:             types.BoolValue(generateNames.Success),
		ResourceNameDetails: types.MapValueMust(types.StringType, detailsMap),
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state", map[string]interface{}{
			"diagnostics": resp.Diagnostics,
		})
		return
	}

	tflog.Info(ctx, "Successfully created generate_name resource", map[string]interface{}{
		"id":            state.ID.ValueInt64(),
		"resource_name": state.ResourceName.ValueString(),
		"success":       state.Success.ValueBool(),
	})
}

// Read refreshes the Terraform state with the latest data.
func (r *generateName) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Starting read operation for generate_name resource")

	// Retrieve current state
	var state generateNameModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to retrieve current state", map[string]interface{}{
			"diagnostics": resp.Diagnostics,
		})
		return
	}

	tflog.Debug(ctx, "Reading generate_name resource", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	// Fetch latest data from API using state.ID
	apiResp, err := r.client.GetName(int16(state.ID.ValueInt64()))
	if err != nil {
		tflog.Error(ctx, "API call to get name failed", map[string]interface{}{
			"error": err.Error(),
			"id":    state.ID.ValueInt64(),
		})

		// Enhanced error handling for Read operation
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			tflog.Warn(ctx, "Resource not found, removing from state", map[string]interface{}{
				"id": state.ID.ValueInt64(),
			})
			resp.State.RemoveResource(ctx)
			return
		} else if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "401") {
			resp.Diagnostics.AddError(
				"Authentication Failed",
				"Invalid API key or insufficient permissions. Please check your API key configuration.",
			)
		} else if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "403") {
			resp.Diagnostics.AddError(
				"Access Forbidden",
				"The API key does not have permission to read this resource.",
			)
		} else if strings.Contains(err.Error(), "timeout") {
			resp.Diagnostics.AddError(
				"Request Timeout",
				"The request to the naming tool API timed out. Please try again.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Read azurenamingtool generated_name",
				fmt.Sprintf("An unexpected error occurred while reading the resource: %s", err.Error()),
			)
		}
		return
	}

	detailsMap := map[string]attr.Value{}
	if apiResp.ResourceTypeName != "" {
		detailsMap["resource_type_name"] = types.StringValue(apiResp.ResourceTypeName)
	}

	// Only update fields that are returned by the API
	state.ResourceName = types.StringValue(apiResp.ResourceName)
	state.ResourceTypeName = types.StringValue(apiResp.ResourceTypeName)
	state.ResourceNameDetails = types.MapValueMust(types.StringType, detailsMap)

	tflog.Debug(ctx, "Successfully retrieved resource data from API", map[string]interface{}{
		"id":            state.ID.ValueInt64(),
		"resource_name": apiResp.ResourceName,
	})

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state during read", map[string]interface{}{
			"diagnostics": resp.Diagnostics,
		})
		return
	}

	tflog.Debug(ctx, "Successfully completed read operation for generate_name resource")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *generateName) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *generateName) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete operation called for generate_name resource")
	tflog.Info(ctx, "Delete operation is currently a no-op - resource will be removed from state only")

	// NOTE: The Azure Naming Tool API currently doesn't support deletion via API key authentication.
	// Delete functionality requires a plain text password in addition to the API key, which is
	// not secure and not supported by this provider. An issue has been raised with the API
	// developer to support deletion using only the API key for authentication.
	//
	// For now, resources are only removed from Terraform state but remain in the naming tool.
	// This means:
	// 1. The same name configuration cannot be recreated
	// 2. Resources accumulate in the naming tool over time
	// 3. Manual cleanup may be required in the naming tool interface
	//
	// TODO: Implement actual deletion when the API supports API-key-only authentication
}

// Configure adds the provider configured client to the data source.
func (r *generateName) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring generate_name resource")

	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		tflog.Debug(ctx, "ProviderData is nil, skipping configuration")
		return
	}

	client, ok := req.ProviderData.(*azurenamingtool.Client)
	if !ok {
		tflog.Error(ctx, "Unexpected provider data type", map[string]interface{}{
			"expected": "*azurenamingtool.Client",
			"actual":   fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError(
			"Unexpected Resource Source Configure Type",
			fmt.Sprintf("Expected *azurenamingtool.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
	tflog.Debug(ctx, "Successfully configured generate_name resource with API client")
}

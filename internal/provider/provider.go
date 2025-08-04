package provider

import (
  "context"
  "os"

  "github.com/ofbjansen/azurenamingtool-client-go"
  "github.com/hashicorp/terraform-plugin-framework/datasource"
  "github.com/hashicorp/terraform-plugin-framework/path"
  "github.com/hashicorp/terraform-plugin-framework/provider"
  "github.com/hashicorp/terraform-plugin-framework/provider/schema"
  "github.com/hashicorp/terraform-plugin-framework/resource"
  "github.com/hashicorp/terraform-plugin-framework/types"
)


// Ensure the implementation satisfies the expected interfaces.
var (
    _ provider.Provider = &proactnamingProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &proactnamingProvider{
            version: version,
        }
    }
}

// proactnamingProvider is the provider implementation.
type proactnamingProvider struct {
    // version is set to the provider version on release, "dev" when the
    // provider is built and ran locally, and "test" when running acceptance
    // testing.
    version string
}

// proactnamingProviderModel maps provider schema data to a Go type.
type proactnamingProviderModel struct {
    Host     types.String `tfsdk:"host"`
    APIKey types.String `tfsdk:"apikey"`
}

// Metadata returns the provider type name.
func (p *proactnamingProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "proactnaming"
    resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *proactnamingProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "host": schema.StringAttribute{
                Optional: true,
            },
            "apikey": schema.StringAttribute{
                Optional: true,
            },
        },
    }
}


func (p *proactnamingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Retrieve provider data from configuration
    var config proactnamingProviderModel
    diags := req.Config.Get(ctx, &config)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // If practitioner provided a configuration value for any of the
    // attributes, it must be a known value.

    if config.Host.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Unknown proactnaming API Host",
            "The provider cannot create the proactnaming API client as there is an unknown configuration value for the proactnaming API host. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the proactnaming_HOST environment variable.",
        )
    }

    if config.APIKey.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("apikey"),
            "Unknown proactnaming API APIKey",
            "The provider cannot create the proactnaming API client as there is an unknown configuration value for the proactnaming API Key. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the APIKEY environment variable.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Default values to environment variables, but override
    // with Terraform configuration value if set.

    host := os.Getenv("PROACTNAMING_HOST")
    apikey := os.Getenv("PROACTNAMING_APIKEY")

    if !config.Host.IsNull() {
        host = config.Host.ValueString()
    }

    if !config.APIKey.IsNull() {
        apikey = config.APIKey.ValueString()
    }

    // If any of the expected configurations are missing, return
    // errors with provider-specific guidance.

    if host == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Missing proactnaming API Host",
            "The provider cannot create the proactnaming API client as there is a missing or empty value for the proactnaming API host. "+
                "Set the host value in the configuration or use the proactnaming_HOST environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if apikey == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("apikey"),
            "Missing proactnaming API API Key",
            "The provider cannot create the proactnaming API client as there is a missing or empty value for the proactnaming API Key. "+
                "Set the apikey value in the configuration or use the proactnaming_APIKEY environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Create a new proactnaming client using the configuration values
    client, err := azurenamingtool.NewClient(&host, &apikey)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create proactnaming API Client",
            "An unexpected error occurred when creating the proactnaming API client. "+
                "If the error is not clear, please contact the provider developers.\n\n"+
                "proactnaming Client Error: "+err.Error(),
        )
        return
    }

    // Make the proactnaming client available during DataSource and Resource
    // type Configure methods.
    resp.DataSourceData = client
    resp.ResourceData = client
}


// DataSources defines the data sources implemented in the provider.
func (p *proactnamingProvider) DataSources(_ context.Context) []func() datasource.DataSource {
  return []func() datasource.DataSource {
	NewresourceTypesDataSource,
  }
}


// Resources defines the resources implemented in the provider.
func (p *proactnamingProvider) Resources(_ context.Context) []func() resource.Resource {
    return nil
}

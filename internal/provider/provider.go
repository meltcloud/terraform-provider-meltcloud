package provider

import (
	"context"
	"os"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure MeltcloudProvider satisfies various provider interfaces.
var _ provider.Provider = &MeltcloudProvider{}
var _ provider.ProviderWithFunctions = &MeltcloudProvider{}

// MeltcloudProvider defines the provider implementation.
type MeltcloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MeltcloudProviderModel describes the provider data model.
type MeltcloudProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	Organization types.String `tfsdk:"organization"`
	APIKey       types.String `tfsdk:"api_key"`
}

func (p *MeltcloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "meltcloud"
	resp.Version = p.version
}

func (p *MeltcloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "URL of the meltcloud API, defaults to https://app.meltcloud.io. Can also be set via MELTCLOUD_ENDPOINT environment variable.",
				Optional:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "UUID of the meltcloud Organization. Can also be set via MELTCLOUD_ORGANIZATION environment variable.",
				Required:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API Key permitted for the organization. Can also be set via MELTCLOUD_API_KEY environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *MeltcloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MeltcloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var apiKey string
	if data.APIKey.IsNull() {
		var found bool
		if apiKey, found = os.LookupEnv("MELTCLOUD_API_KEY"); !found {
			resp.Diagnostics.AddError("Config Error", "either MELTCLOUD_API_KEY or api_key in provider config must be set")
			return
		}
	} else {
		apiKey = data.APIKey.ValueString()
	}

	var endpoint string
	if data.Endpoint.IsNull() {
		var found bool
		if endpoint, found = os.LookupEnv("MELTCLOUD_ENDPOINT"); !found {
			endpoint = "https://app.meltcloud.io"
		}
	} else {
		endpoint = data.Endpoint.ValueString()
	}

	var organization string
	if data.Organization.IsNull() {
		var found bool
		if organization, found = os.LookupEnv("MELTCLOUD_ORGANIZATION"); !found {
			resp.Diagnostics.AddError("Config Error", "either MELTCLOUD_ORGANIZATION or organization in provider config must be set")
			return
		}
	} else {
		organization = data.Organization.ValueString()
	}

	apiClient := client.New(endpoint, organization, apiKey)
	apiClient.HttpClient.Debug = true
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *MeltcloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewMachinePoolResource,
		NewMachineResource,
		NewEnrollmentImageResource,
		NewIPXEBootArtifactResource,
		NewIPXEChainURLResource,
		NewUEFIHTTPBootURLResource,
		NewNetworkProfileResource,
	}
}

func (p *MeltcloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewClusterDataSource,
		NewMachinePoolDataSource,
		NewMachineDataSource,
		NewEnrollmentImageDataSource,
		NewIPXEBootArtifactDataSource,
		NewIPXEChainURLDataSource,
		NewUEFIHTTPBootURLDataSource,
		NewNetworkProfileDataSource,
	}
}

func (p *MeltcloudProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewCustomizeUUIDInIPXEScriptFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MeltcloudProvider{
			version: version,
		}
	}
}

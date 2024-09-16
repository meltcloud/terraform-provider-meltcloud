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

// Ensure MeltProvider satisfies various provider interfaces.
var _ provider.Provider = &MeltProvider{}
var _ provider.ProviderWithFunctions = &MeltProvider{}

// MeltProvider defines the provider implementation.
type MeltProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MeltProviderModel describes the provider data model.
type MeltProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	Organization types.String `tfsdk:"organization"`
	APIKey       types.String `tfsdk:"api_key"`
}

func (p *MeltProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "melt"
	resp.Version = p.version
}

func (p *MeltProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "URL of the Melt API, i.e. https://api.meltcloud.io",
				Optional:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "UUID of the Melt Organization",
				Required:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API Key permitted for the organization. Can also be set via MELT_API_KEY environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *MeltProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MeltProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var apiKey string
	if data.APIKey.IsNull() {
		var found bool
		if apiKey, found = os.LookupEnv("MELT_API_KEY"); !found {
			resp.Diagnostics.AddError("Config Error", "either MELT_API_KEY or api_key in provider config must be set")
			return
		}
	} else {
		apiKey = data.APIKey.ValueString()
	}

	// TODO validate somehow?

	// Example client configuration for data sources and resources
	client := client.New(data.Endpoint.ValueString(), data.Organization.ValueString(), apiKey)
	client.HttpClient.Debug = true
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MeltProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewMachinePoolResource,
		NewMachineResource,
		NewIPXEBootArtifactResource,
		NewIPXEChainURLResource,
	}
}

func (p *MeltProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMachineDataSource,
	}
}

func (p *MeltProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MeltProvider{
			version: version,
		}
	}
}

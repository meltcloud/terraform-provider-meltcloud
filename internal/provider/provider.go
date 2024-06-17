package provider

import (
	"context"
	"terraform-provider-melt/internal/client"

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
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *MeltProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "melt"
	resp.Version = p.version
}

func (p *MeltProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
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

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := client.New()
	client.SetBaseURL(data.Endpoint.ValueString())
	client.HttpClient.Debug = true
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MeltProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewMachinePoolResource,
		NewMachineResource,
	}
}

func (p *MeltProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMachineDataSource,
	}
}

func (p *MeltProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MeltProvider{
			version: version,
		}
	}
}

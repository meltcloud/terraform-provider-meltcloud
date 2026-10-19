package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FoundryDataSource{}

func NewFoundryDataSource() datasource.DataSource {
	return &FoundryDataSource{}
}

type FoundryDataSource struct {
	client *client.Client
}

type FoundryDataSourceModel struct {
	ID                types.Int64                        `tfsdk:"id"`
	Status            types.String                       `tfsdk:"status"`
	Version           types.String                       `tfsdk:"version"`
	HATopology        types.String                       `tfsdk:"ha_topology"`
	RailsLogLevel     types.String                       `tfsdk:"rails_log_level"`
	IdentityProvider  types.String                       `tfsdk:"identity_provider"`
	Auth0Config       *FoundryDataSourceAuth0Model       `tfsdk:"auth0_config"`
	LetsencryptConfig *FoundryDataSourceLetsencryptModel `tfsdk:"letsencrypt_config"`
	SentryConfig      *FoundryDataSourceSentryModel      `tfsdk:"sentry_config"`
}

type FoundryDataSourceAuth0Model struct {
	Domain   types.String `tfsdk:"domain"`
	ClientID types.String `tfsdk:"client_id"`
}

type FoundryDataSourceLetsencryptModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Email   types.String `tfsdk:"email"`
	Staging types.Bool   `tfsdk:"staging"`
}

type FoundryDataSourceSentryModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Environment types.String `tfsdk:"environment"`
}

func (d *FoundryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_foundry"
}

func (d *FoundryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resource := foundryResourceAttributes()
	description := func(name string) string { return resource[name].GetMarkdownDescription() }
	str := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, MarkdownDescription: description}
	}
	boolean := func(description string) schema.BoolAttribute {
		return schema.BoolAttribute{Computed: true, MarkdownDescription: description}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "The [Foundry](https://docs.meltcloud.io/concepts/nest/overview#foundry), the meltcloud management plane that runs on the Nest. " +
			"Needs the API Key of the admin Organization. The secrets are never returned.",
		Attributes: map[string]schema.Attribute{
			"id":                schema.Int64Attribute{Computed: true, MarkdownDescription: description("id")},
			"status":            str(description("status")),
			"version":           str("The Foundry version it runs"),
			"ha_topology":       str(description("ha_topology")),
			"rails_log_level":   str(description("rails_log_level")),
			"identity_provider": str(description("identity_provider")),
			"auth0_config": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The Auth0 tenant users sign in through. Null when none is set",
				Attributes: map[string]schema.Attribute{
					"domain":    str("Domain of the Auth0 tenant"),
					"client_id": str("Client ID of the Auth0 application"),
				},
			},
			"letsencrypt_config": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Certificates from Let's Encrypt",
				Attributes: map[string]schema.Attribute{
					"enabled": boolean("Whether the Foundry hostnames get Let's Encrypt certificates"),
					"email":   str("Email address Let's Encrypt sends expiry notices to"),
					"staging": boolean("Whether the Let's Encrypt staging environment is used"),
				},
			},
			"sentry_config": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Error reporting to Sentry",
				Attributes: map[string]schema.Attribute{
					"enabled":     boolean("Whether errors of the Foundry application are reported to Sentry"),
					"environment": str("Environment the errors are reported under"),
				},
			},
		},
	}
}

func (d *FoundryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *FoundryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Foundry().Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Foundry, got error: %s", err))
		return
	}

	foundry := result.Foundry
	data := FoundryDataSourceModel{
		ID:               types.Int64Value(foundry.ID),
		Status:           types.StringValue(foundry.Status),
		Version:          types.StringValue(foundry.Version),
		HATopology:       types.StringValue(foundry.HATopology),
		RailsLogLevel:    types.StringValue(foundry.RailsLogLevel),
		IdentityProvider: types.StringValue(foundry.IdentityProvider),
		LetsencryptConfig: &FoundryDataSourceLetsencryptModel{
			Enabled: types.BoolValue(foundry.LetsencryptEnabled),
			Email:   emptyAsNull(foundry.LetsencryptEmail),
			Staging: types.BoolValue(foundry.LetsencryptStaging),
		},
		SentryConfig: &FoundryDataSourceSentryModel{
			Enabled:     types.BoolValue(foundry.SentryEnabled),
			Environment: emptyAsNull(foundry.SentryEnvironment),
		},
	}
	if foundry.Auth0Domain != "" {
		data.Auth0Config = &FoundryDataSourceAuth0Model{
			Domain:   types.StringValue(foundry.Auth0Domain),
			ClientID: types.StringValue(foundry.Auth0ClientID),
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

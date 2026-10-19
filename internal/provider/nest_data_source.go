package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NestDataSource{}

func NewNestDataSource() datasource.DataSource {
	return &NestDataSource{}
}

type NestDataSource struct {
	client *client.Client
}

func (d *NestDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest"
}

func (d *NestDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	computed := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, MarkdownDescription: description}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "The [Nest](https://docs.meltcloud.io/concepts/nest/overview), the management appliance that hosts the Foundry and the Kubernetes control planes. " +
			"Needs the API Key of the admin Organization.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.Int64Attribute{Computed: true, MarkdownDescription: "Internal ID of the Nest on meltcloud"},
			"name":    computed("Name of the Nest, given at install"),
			"status":  computed("Status of the Nest, for example `ready` or `updating`"),
			"version": computed("The Nest version it runs"),
		},
	}
}

func (d *NestDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *NestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Nest().Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest, got error: %s", err))
		return
	}

	nest := result.Nest
	resp.Diagnostics.Append(resp.State.Set(ctx, &NestResourceModel{
		ID:      types.Int64Value(nest.ID),
		Name:    types.StringValue(nest.Name),
		Status:  types.StringValue(nest.Status),
		Version: types.StringValue(nest.Version),
	})...)
}

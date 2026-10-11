package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ZoneDataSource{}

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

type ZoneDataSource struct {
	client *client.Client
}

type ZoneDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	RegionID types.Int64  `tfsdk:"region_id"`
	Name     types.String `tfsdk:"name"`
}

func (d *ZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := zoneResourceAttributes()
	resp.Schema = schema.Schema{
		MarkdownDescription: zoneDesc,
		Attributes: withLookupAttributes(map[string]schema.Attribute{
			"region_id": schema.Int64Attribute{
				MarkdownDescription: attributes["region_id"].GetMarkdownDescription(),
				Required:            true,
			},
		}, attributes["id"].GetMarkdownDescription(), attributes["name"].GetMarkdownDescription()),
	}
}

func (d *ZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ZoneDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *client.ZoneResult
	var err *client.Error
	if data.ID.IsNull() {
		result, err = d.client.Zone().GetByName(ctx, data.RegionID.ValueInt64(), data.Name.ValueString())
	} else {
		result, err = d.client.Zone().Get(ctx, data.RegionID.ValueInt64(), data.ID.ValueInt64())
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Zone, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.Zone.ID)
	data.Name = types.StringValue(result.Zone.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

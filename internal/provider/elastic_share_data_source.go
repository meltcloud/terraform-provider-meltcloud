package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ElasticShareDataSource{}

func NewElasticShareDataSource() datasource.DataSource {
	return &ElasticShareDataSource{}
}

type ElasticShareDataSource struct {
	client *client.Client
}

type ElasticShareDataSourceModel struct {
	ID                        types.Int64  `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Cores                     types.Int64  `tfsdk:"cores"`
	DiskGB                    types.Int64  `tfsdk:"disk_gb"`
	MemoryMB                  types.Int64  `tfsdk:"memory_mb"`
	CapacityID                types.Int64  `tfsdk:"capacity_id"`
	ConsumingOrganizationUUID types.String `tfsdk:"consuming_organization_uuid"`
}

func (d *ElasticShareDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_share"
}

func (d *ElasticShareDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticShareDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: elasticShareResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: elasticShareResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: elasticShareResourceAttributes()["cores"].GetMarkdownDescription(),
				Computed:            true,
			},
			"disk_gb": schema.Int64Attribute{
				MarkdownDescription: elasticShareResourceAttributes()["disk_gb"].GetMarkdownDescription(),
				Computed:            true,
			},
			"memory_mb": schema.Int64Attribute{
				MarkdownDescription: elasticShareResourceAttributes()["memory_mb"].GetMarkdownDescription(),
				Computed:            true,
			},
			"capacity_id": schema.Int64Attribute{
				MarkdownDescription: elasticShareResourceAttributes()["capacity_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"consuming_organization_uuid": schema.StringAttribute{
				MarkdownDescription: elasticShareResourceAttributes()["consuming_organization_uuid"].GetMarkdownDescription(),
				Computed:            true,
			},
		},
	}
}

func (d *ElasticShareDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *ElasticShareDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticShareDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ElasticShare().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic share by ID %d, got error: %s", data.ID.ValueInt64(), err))
		return
	}
	share := result.ElasticShare

	data.ID = types.Int64Value(share.ID)
	data.Name = types.StringValue(share.Name)
	data.Cores = types.Int64Value(share.Cores)
	data.DiskGB = types.Int64Value(share.DiskGB)
	data.MemoryMB = types.Int64Value(share.MemoryMB)
	data.CapacityID = types.Int64Value(share.CapacityID)
	data.ConsumingOrganizationUUID = types.StringValue(share.ConsumingOrganizationUUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

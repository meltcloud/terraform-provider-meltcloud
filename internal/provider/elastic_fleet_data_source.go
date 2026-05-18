package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ElasticFleetDataSource{}

func NewElasticFleetDataSource() datasource.DataSource {
	return &ElasticFleetDataSource{}
}

type ElasticFleetDataSource struct {
	client *client.Client
}

type ElasticFleetDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ClusterID types.Int64  `tfsdk:"cluster_id"`
	Status    types.String `tfsdk:"status"`
}

func (d *ElasticFleetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_fleet"
}

func (d *ElasticFleetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticFleetDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: elasticFleetResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: elasticFleetResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"cluster_id": schema.Int64Attribute{
				MarkdownDescription: elasticFleetResourceAttributes()["cluster_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: elasticFleetResourceAttributes()["status"].GetMarkdownDescription(),
				Computed:            true,
			},
		},
	}
}

func (d *ElasticFleetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ElasticFleetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticFleetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ElasticFleet().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic fleet with ID %d, got error: %s", data.ID.ValueInt64(), err))
		return
	}

	data.ID = types.Int64Value(result.ElasticFleet.ID)
	data.Name = types.StringValue(result.ElasticFleet.Name)
	data.ClusterID = types.Int64Value(result.ElasticFleet.ClusterID)
	data.Status = types.StringValue(result.ElasticFleet.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

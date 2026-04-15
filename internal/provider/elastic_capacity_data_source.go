package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ElasticCapacityDataSource{}

func NewElasticCapacityDataSource() datasource.DataSource {
	return &ElasticCapacityDataSource{}
}

type ElasticCapacityDataSource struct {
	client *client.Client
}

type ElasticCapacityDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ClusterID types.Int64  `tfsdk:"cluster_id"`
	Status    types.String `tfsdk:"status"`
}

func (d *ElasticCapacityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_capacity"
}

func (d *ElasticCapacityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticCapacityDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: elasticCapacityResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: elasticCapacityResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"cluster_id": schema.Int64Attribute{
				MarkdownDescription: elasticCapacityResourceAttributes()["cluster_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: elasticCapacityResourceAttributes()["status"].GetMarkdownDescription(),
				Computed:            true,
			},
		},
	}
}

func (d *ElasticCapacityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ElasticCapacityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticCapacityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ElasticCapacity().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic capacity with ID %d, got error: %s", data.ID.ValueInt64(), err))
		return
	}

	data.ID = types.Int64Value(result.ElasticCapacity.ID)
	data.Name = types.StringValue(result.ElasticCapacity.Name)
	data.ClusterID = types.Int64Value(result.ElasticCapacity.ClusterID)
	data.Status = types.StringValue(result.ElasticCapacity.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

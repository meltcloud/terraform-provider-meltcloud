package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ElasticNodePoolDataSource{}

func NewElasticNodePoolDataSource() datasource.DataSource {
	return &ElasticNodePoolDataSource{}
}

type ElasticNodePoolDataSource struct {
	client *client.Client
}

type ElasticNodePoolDataSourceModel struct {
	ID             types.Int64                `tfsdk:"id"`
	ClusterID      types.Int64                `tfsdk:"cluster_id"`
	Name           types.String               `tfsdk:"name"`
	ElasticQuotaID types.Int64                `tfsdk:"elastic_quota_id"`
	Version        types.String               `tfsdk:"version"`
	PatchVersion   types.String               `tfsdk:"patch_version"`
	NodeCount      types.Int64                `tfsdk:"node_count"`
	Status         types.String               `tfsdk:"status"`
	NodeConfig     *NodeConfigDataSourceModel `tfsdk:"node_config"`
}

type NodeConfigDataSourceModel struct {
	Cores    types.Int64 `tfsdk:"cores"`
	MemoryMB types.Int64 `tfsdk:"memory_mb"`
	DiskGB   types.Int64 `tfsdk:"disk_gb"`
}

func (d *ElasticNodePoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_node_pool"
}

func (d *ElasticNodePoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticNodePoolDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"cluster_id": schema.Int64Attribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["cluster_id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"elastic_quota_id": schema.Int64Attribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["elastic_quota_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"patch_version": schema.StringAttribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["patch_version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"node_count": schema.Int64Attribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["node_count"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: elasticNodePoolResourceAttributes()["status"].GetMarkdownDescription(),
				Computed:            true,
			},
			"node_config": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Per-node resource configuration",
				Attributes: map[string]schema.Attribute{
					"cores": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Number of cores per node",
					},
					"memory_mb": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Memory in MB per node",
					},
					"disk_gb": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Disk in GB per node",
					},
				},
			},
		},
	}
}

func (d *ElasticNodePoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ElasticNodePoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticNodePoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ElasticNodePool().Get(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic node pool with ID %d on cluster ID %d, got error: %s", data.ID.ValueInt64(), data.ClusterID.ValueInt64(), err))
		return
	}

	data.ID = types.Int64Value(result.ElasticNodePool.ID)
	data.ClusterID = types.Int64Value(result.ElasticNodePool.ClusterID)
	data.Name = types.StringValue(result.ElasticNodePool.Name)
	data.ElasticQuotaID = types.Int64Value(result.ElasticNodePool.ElasticQuotaID)
	data.Version = types.StringValue(result.ElasticNodePool.Version)
	data.PatchVersion = types.StringValue(result.ElasticNodePool.PatchVersion)
	data.NodeCount = types.Int64Value(result.ElasticNodePool.NodeCount)
	data.Status = types.StringValue(result.ElasticNodePool.Status)
	data.NodeConfig = &NodeConfigDataSourceModel{
		Cores:    types.Int64Value(result.ElasticNodePool.NodeCores),
		MemoryMB: types.Int64Value(result.ElasticNodePool.NodeMemoryMB),
		DiskGB:   types.Int64Value(result.ElasticNodePool.NodeDiskGB),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

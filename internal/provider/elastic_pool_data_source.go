package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ElasticPoolDataSource{}

func NewElasticPoolDataSource() datasource.DataSource {
	return &ElasticPoolDataSource{}
}

type ElasticPoolDataSource struct {
	client *client.Client
}

type ElasticPoolDataSourceModel struct {
	ID           types.Int64                `tfsdk:"id"`
	ClusterID    types.Int64                `tfsdk:"cluster_id"`
	Name         types.String               `tfsdk:"name"`
	ShareID      types.Int64                `tfsdk:"share_id"`
	Version      types.String               `tfsdk:"version"`
	PatchVersion types.String               `tfsdk:"patch_version"`
	NodeCount    types.Int64                `tfsdk:"node_count"`
	Status       types.String               `tfsdk:"status"`
	NodeConfig   *NodeConfigDataSourceModel `tfsdk:"node_config"`
}

type NodeConfigDataSourceModel struct {
	Cores    types.Int64 `tfsdk:"cores"`
	MemoryMB types.Int64 `tfsdk:"memory_mb"`
	DiskGB   types.Int64 `tfsdk:"disk_gb"`
}

func (d *ElasticPoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_pool"
}

func (d *ElasticPoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticPoolDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: elasticPoolResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"cluster_id": schema.Int64Attribute{
				MarkdownDescription: elasticPoolResourceAttributes()["cluster_id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: elasticPoolResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"share_id": schema.Int64Attribute{
				MarkdownDescription: elasticPoolResourceAttributes()["share_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: elasticPoolResourceAttributes()["version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"patch_version": schema.StringAttribute{
				MarkdownDescription: elasticPoolResourceAttributes()["patch_version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"node_count": schema.Int64Attribute{
				MarkdownDescription: elasticPoolResourceAttributes()["node_count"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: elasticPoolResourceAttributes()["status"].GetMarkdownDescription(),
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

func (d *ElasticPoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ElasticPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ElasticPool().Get(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic pool with ID %d on cluster ID %d, got error: %s", data.ID.ValueInt64(), data.ClusterID.ValueInt64(), err))
		return
	}

	data.ID = types.Int64Value(result.ElasticPool.ID)
	data.ClusterID = types.Int64Value(result.ElasticPool.ClusterID)
	data.Name = types.StringValue(result.ElasticPool.Name)
	data.ShareID = types.Int64Value(result.ElasticPool.ShareID)
	data.Version = types.StringValue(result.ElasticPool.Version)
	data.PatchVersion = types.StringValue(result.ElasticPool.PatchVersion)
	data.NodeCount = types.Int64Value(result.ElasticPool.NodeCount)
	data.Status = types.StringValue(result.ElasticPool.Status)
	data.NodeConfig = &NodeConfigDataSourceModel{
		Cores:    types.Int64Value(result.ElasticPool.NodeCores),
		MemoryMB: types.Int64Value(result.ElasticPool.NodeMemoryMB),
		DiskGB:   types.Int64Value(result.ElasticPool.NodeDiskGB),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

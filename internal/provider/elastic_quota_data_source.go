package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ElasticQuotaDataSource{}

func NewElasticQuotaDataSource() datasource.DataSource {
	return &ElasticQuotaDataSource{}
}

type ElasticQuotaDataSource struct {
	client *client.Client
}

type ElasticQuotaDataSourceModel struct {
	ID                        types.Int64  `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	VCPUs                     types.Int64  `tfsdk:"vcpus"`
	DiskGiB                   types.Int64  `tfsdk:"disk_gib"`
	MemoryMiB                 types.Int64  `tfsdk:"memory_mib"`
	ElasticFleetID            types.Int64  `tfsdk:"elastic_fleet_id"`
	ConsumingOrganizationUUID types.String `tfsdk:"consuming_organization_uuid"`
}

func (d *ElasticQuotaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_quota"
}

func (d *ElasticQuotaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticQuotaDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"vcpus": schema.Int64Attribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["vcpus"].GetMarkdownDescription(),
				Computed:            true,
			},
			"disk_gib": schema.Int64Attribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["disk_gib"].GetMarkdownDescription(),
				Computed:            true,
			},
			"memory_mib": schema.Int64Attribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["memory_mib"].GetMarkdownDescription(),
				Computed:            true,
			},
			"elastic_fleet_id": schema.Int64Attribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["elastic_fleet_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"consuming_organization_uuid": schema.StringAttribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["consuming_organization_uuid"].GetMarkdownDescription(),
				Computed:            true,
			},
		},
	}
}

func (d *ElasticQuotaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ElasticQuotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticQuotaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ElasticQuota().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic quota by ID %d, got error: %s", data.ID.ValueInt64(), err))
		return
	}
	quota := result.ElasticQuota

	data.ID = types.Int64Value(quota.ID)
	data.Name = types.StringValue(quota.Name)
	data.VCPUs = types.Int64Value(quota.VCPUs)
	data.DiskGiB = types.Int64Value(quota.DiskGiB)
	data.MemoryMiB = types.Int64Value(quota.MemoryMiB)
	data.ElasticFleetID = types.Int64Value(quota.ElasticFleetID)
	data.ConsumingOrganizationUUID = types.StringValue(quota.ConsumingOrganizationUUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

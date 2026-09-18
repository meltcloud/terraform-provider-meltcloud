package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		Attributes: withLookupAttributes(map[string]schema.Attribute{
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
				MarkdownDescription: elasticQuotaResourceAttributes()["elastic_fleet_id"].GetMarkdownDescription() + ", required with `name`",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"consuming_organization_uuid": schema.StringAttribute{
				MarkdownDescription: elasticQuotaResourceAttributes()["consuming_organization_uuid"].GetMarkdownDescription(),
				Computed:            true,
			},
		}, elasticQuotaResourceAttributes()["id"].GetMarkdownDescription(), elasticQuotaResourceAttributes()["name"].GetMarkdownDescription()),
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

func (d *ElasticQuotaDataSource) readQuota(ctx context.Context, data ElasticQuotaDataSourceModel) (*client.ElasticQuotaResult, *client.Error) {
	if data.ID.IsNull() {
		return d.client.ElasticQuota().GetByName(ctx, data.ElasticFleetID.ValueInt64(), data.Name.ValueString())
	}
	return d.client.ElasticQuota().Get(ctx, data.ID.ValueInt64())
}

func (d *ElasticQuotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticQuotaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() && data.ElasticFleetID.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("elastic_fleet_id"), "Missing Attribute", "A quota's name is unique within its fleet, so a lookup by name needs elastic_fleet_id.")
		return
	}

	result, err := d.readQuota(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic quota, got error: %s", err))
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

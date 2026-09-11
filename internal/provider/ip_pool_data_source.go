package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IPPoolDataSource{}

func NewIPPoolDataSource() datasource.DataSource {
	return &IPPoolDataSource{}
}

type IPPoolDataSource struct {
	client *client.Client
}

type IPPoolDataSourceModel struct {
	ID          types.Int64                  `tfsdk:"id"`
	Name        types.String                 `tfsdk:"name"`
	CIDR        types.String                 `tfsdk:"cidr"`
	Description types.String                 `tfsdk:"description"`
	Ranges      []IPPoolRangeDataSourceModel `tfsdk:"ranges"`
}

type IPPoolRangeDataSourceModel struct {
	Kind         types.String `tfsdk:"kind"`
	StartAddress types.String `tfsdk:"start_address"`
	EndAddress   types.String `tfsdk:"end_address"`
	Description  types.String `tfsdk:"description"`
}

func (d *IPPoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_pool"
}

func (d *IPPoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := ipPoolResourceAttributes()
	rangeAttributes := ipPoolRangeResourceAttributes()
	resp.Schema = schema.Schema{
		MarkdownDescription: ipPoolDesc,
		Attributes: withLookupAttributes(map[string]schema.Attribute{
			"cidr": schema.StringAttribute{
				MarkdownDescription: attributes["cidr"].GetMarkdownDescription(),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: attributes["description"].GetMarkdownDescription(),
				Computed:            true,
			},
			"ranges": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: rangeAttributes["kind"].GetMarkdownDescription(),
						},
						"start_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: rangeAttributes["start_address"].GetMarkdownDescription(),
						},
						"end_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: rangeAttributes["end_address"].GetMarkdownDescription(),
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: rangeAttributes["description"].GetMarkdownDescription(),
						},
					},
				},
			},
		}, attributes["id"].GetMarkdownDescription(), attributes["name"].GetMarkdownDescription()),
	}
}

func (d *IPPoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IPPoolDataSource) readIPPool(ctx context.Context, data IPPoolDataSourceModel) (*client.IPPoolResult, *client.Error) {
	if data.ID.IsNull() {
		return d.client.IPPool().GetByName(ctx, data.Name.ValueString())
	}
	return d.client.IPPool().Get(ctx, data.ID.ValueInt64())
}

func (d *IPPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.readIPPool(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IP Pool, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.IPPool.ID)
	data.Name = types.StringValue(result.IPPool.Name)
	data.CIDR = types.StringValue(result.IPPool.CIDR)
	data.Description = types.StringPointerValue(result.IPPool.Description)

	for _, poolRange := range result.IPPool.Ranges {
		data.Ranges = append(data.Ranges, IPPoolRangeDataSourceModel{
			Kind:         types.StringValue(poolRange.Kind),
			StartAddress: types.StringValue(poolRange.StartAddress),
			EndAddress:   types.StringValue(poolRange.EndAddress),
			Description:  types.StringPointerValue(poolRange.Description),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SubnetDataSource{}

func NewSubnetDataSource() datasource.DataSource {
	return &SubnetDataSource{}
}

type SubnetDataSource struct {
	client *client.Client
}

type SubnetDataSourceModel struct {
	ID         types.Int64            `tfsdk:"id"`
	NetworkID  types.Int64            `tfsdk:"network_id"`
	Name       types.String           `tfsdk:"name"`
	VLAN       types.Int64            `tfsdk:"vlan"`
	Addressing types.String           `tfsdk:"addressing"`
	IPPoolID   types.Int64            `tfsdk:"ip_pool_id"`
	Gateway    types.String           `tfsdk:"gateway"`
	MTU        types.Int64            `tfsdk:"mtu"`
	DNS        types.List             `tfsdk:"dns"`
	NTP        types.List             `tfsdk:"ntp"`
	Domains    types.List             `tfsdk:"domains"`
	Routes     []RouteDataSourceModel `tfsdk:"routes"`
}

type RouteDataSourceModel struct {
	Destination types.String `tfsdk:"destination"`
	Via         types.String `tfsdk:"via"`
	Metric      types.Int64  `tfsdk:"metric"`
}

func (d *SubnetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *SubnetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := subnetResourceAttributes()
	routeAttributes := routeResourceAttributes()
	resp.Schema = schema.Schema{
		MarkdownDescription: subnetDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: attributes["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"network_id": schema.Int64Attribute{
				MarkdownDescription: attributes["network_id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: attributes["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"vlan": schema.Int64Attribute{
				MarkdownDescription: attributes["vlan"].GetMarkdownDescription(),
				Computed:            true,
			},
			"addressing": schema.StringAttribute{
				MarkdownDescription: attributes["addressing"].GetMarkdownDescription(),
				Computed:            true,
			},
			"ip_pool_id": schema.Int64Attribute{
				MarkdownDescription: attributes["ip_pool_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: attributes["gateway"].GetMarkdownDescription(),
				Computed:            true,
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: attributes["mtu"].GetMarkdownDescription(),
				Computed:            true,
			},
			"dns": schema.ListAttribute{
				MarkdownDescription: attributes["dns"].GetMarkdownDescription(),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"ntp": schema.ListAttribute{
				MarkdownDescription: attributes["ntp"].GetMarkdownDescription(),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"domains": schema.ListAttribute{
				MarkdownDescription: attributes["domains"].GetMarkdownDescription(),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"routes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"destination": schema.StringAttribute{
							MarkdownDescription: routeAttributes["destination"].GetMarkdownDescription(),
							Computed:            true,
						},
						"via": schema.StringAttribute{
							MarkdownDescription: routeAttributes["via"].GetMarkdownDescription(),
							Computed:            true,
						},
						"metric": schema.Int64Attribute{
							MarkdownDescription: routeAttributes["metric"].GetMarkdownDescription(),
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SubnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.Subnet().Get(ctx, data.NetworkID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read subnet with ID %d, got error: %s", data.ID.ValueInt64(), err))
		return
	}

	subnet := result.Subnet
	data.Name = types.StringValue(subnet.Name)
	data.VLAN = types.Int64PointerValue(subnet.VLAN)
	data.Addressing = types.StringValue(subnet.Addressing)
	data.IPPoolID = types.Int64PointerValue(subnet.IPPoolID)
	data.Gateway = types.StringPointerValue(subnet.Gateway)
	data.MTU = types.Int64PointerValue(subnet.MTU)

	dns, diags := stringListValue(ctx, subnet.DNS)
	resp.Diagnostics.Append(diags...)
	ntp, diags := stringListValue(ctx, subnet.NTP)
	resp.Diagnostics.Append(diags...)
	domains, diags := stringListValue(ctx, subnet.Domains)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.DNS = dns
	data.NTP = ntp
	data.Domains = domains

	data.Routes = nil
	for _, route := range subnet.Routes {
		data.Routes = append(data.Routes, RouteDataSourceModel{
			Destination: types.StringValue(route.Destination),
			Via:         types.StringValue(route.Via),
			Metric:      types.Int64PointerValue(route.Metric),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

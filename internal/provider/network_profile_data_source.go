package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NetworkProfileDataSource{}

func NewNetworkProfileDataSource() datasource.DataSource {
	return &NetworkProfileDataSource{}
}

// NetworkProfileDataSource defines the data source implementation.
type NetworkProfileDataSource struct {
	client *client.Client
}

// NetworkProfileDataSourceModel describes the data source data model.
type NetworkProfileDataSourceModel struct {
	ID      types.Int64             `tfsdk:"id"`
	Name    types.String            `tfsdk:"name"`
	Status  types.String            `tfsdk:"status"`
	Uplinks []UplinkDataSourceModel `tfsdk:"uplinks"`
}

type UplinkDataSourceModel struct {
	Name         types.String                 `tfsdk:"name"`
	Mode         types.String                 `tfsdk:"mode"`
	Identifier   types.String                 `tfsdk:"identifier"`
	Interfaces   types.List                   `tfsdk:"interfaces"`
	LACP         types.Bool                   `tfsdk:"lacp"`
	HostNetworks []HostNetworkDataSourceModel `tfsdk:"host_networks"`
}

type HostNetworkDataSourceModel struct {
	SubnetID   types.Int64 `tfsdk:"subnet_id"`
	VLANTagged types.Bool  `tfsdk:"vlan_tagged"`
	Primary    types.Bool  `tfsdk:"primary"`
}

func (d *NetworkProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_profile"
}

func (d *NetworkProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: networkProfileDesc,

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: networkProfileResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: networkProfileResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the Network Profile",
				Computed:            true,
			},
			"uplinks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: uplinkResourceAttributes()["name"].GetMarkdownDescription(),
						},
						"mode": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: uplinkResourceAttributes()["mode"].GetMarkdownDescription(),
						},
						"identifier": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: uplinkResourceAttributes()["identifier"].GetMarkdownDescription(),
						},
						"interfaces": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: uplinkResourceAttributes()["interfaces"].GetMarkdownDescription(),
						},
						"lacp": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: uplinkResourceAttributes()["lacp"].GetMarkdownDescription(),
						},
						"host_networks": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"subnet_id": schema.Int64Attribute{
										Computed:            true,
										MarkdownDescription: hostNetworkResourceAttributes()["subnet_id"].GetMarkdownDescription(),
									},
									"vlan_tagged": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: hostNetworkResourceAttributes()["vlan_tagged"].GetMarkdownDescription(),
									},
									"primary": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: hostNetworkResourceAttributes()["primary"].GetMarkdownDescription(),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *NetworkProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *NetworkProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.NetworkProfile().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network profile with ID %d , got error: %s", data.ID.ValueInt64(), err))
		return
	}

	for _, uplink := range result.NetworkProfile.Uplinks {
		interfacesList, diags := types.ListValueFrom(ctx, types.StringType, uplink.Interfaces)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		var hostNetworks []HostNetworkDataSourceModel
		for _, hostNetwork := range uplink.HostNetworks {
			hostNetworks = append(hostNetworks, HostNetworkDataSourceModel{
				SubnetID:   types.Int64Value(hostNetwork.SubnetID),
				VLANTagged: types.BoolValue(hostNetwork.VLANTagged),
				Primary:    types.BoolValue(hostNetwork.Primary),
			})
		}

		data.Uplinks = append(data.Uplinks, UplinkDataSourceModel{
			Name:         types.StringValue(uplink.Name),
			Mode:         types.StringValue(uplink.Mode),
			Identifier:   types.StringValue(uplink.Identifier),
			Interfaces:   interfacesList,
			LACP:         types.BoolValue(uplink.LACP),
			HostNetworks: hostNetworks,
		})
	}

	data.ID = types.Int64Value(result.NetworkProfile.ID)
	data.Name = types.StringValue(result.NetworkProfile.Name)
	data.Status = types.StringValue(result.NetworkProfile.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

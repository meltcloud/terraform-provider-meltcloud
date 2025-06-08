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
	VLANs   []VLANDataSourceModel   `tfsdk:"vlans"`
	Bridges []BridgeDataSourceModel `tfsdk:"bridges"`
	Bonds   []BondDataSourceModel   `tfsdk:"bonds"`
}

type VLANDataSourceModel struct {
	VLAN      types.Int64  `tfsdk:"vlan"`
	Interface types.String `tfsdk:"interface"`
	DHCP      types.Bool   `tfsdk:"dhcp"`
}

type BridgeDataSourceModel struct {
	Name      types.String `tfsdk:"name"`
	Interface types.String `tfsdk:"interface"`
	DHCP      types.Bool   `tfsdk:"dhcp"`
}

type BondDataSourceModel struct {
	Name       types.String `tfsdk:"name"`
	Interfaces types.String `tfsdk:"interfaces"`
	Kind       types.String `tfsdk:"kind"`
	DHCP       types.Bool   `tfsdk:"dhcp"`
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
			"vlans": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"vlan": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: vlanResourceAttributes()["vlan"].GetMarkdownDescription(),
						},

						"dhcp": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: vlanResourceAttributes()["dhcp"].GetMarkdownDescription(),
						},

						"interface": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: vlanResourceAttributes()["interface"].GetMarkdownDescription(),
						},
					},
				},
			},
			"bridges": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: bridgeResourceAttributes()["name"].GetMarkdownDescription(),
						},
						"interface": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: bridgeResourceAttributes()["interface"].GetMarkdownDescription(),
						},
						"dhcp": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: bridgeResourceAttributes()["dhcp"].GetMarkdownDescription(),
						},
					},
				},
			},
			"bonds": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: bondResourceAttributes()["name"].GetMarkdownDescription(),
						},
						"interfaces": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: bondResourceAttributes()["interfaces"].GetMarkdownDescription(),
						},
						"kind": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: bondResourceAttributes()["kind"].GetMarkdownDescription(),
						},
						"dhcp": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: bondResourceAttributes()["dhcp"].GetMarkdownDescription(),
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

	for _, vlan := range result.NetworkProfile.VLANs {
		data.VLANs = append(data.VLANs, VLANDataSourceModel{
			VLAN:      types.Int64Value(vlan.VLAN),
			DHCP:      types.BoolValue(vlan.DHCP),
			Interface: types.StringValue(vlan.Interface),
		})
	}

	for _, bridge := range result.NetworkProfile.Bridges {
		data.Bridges = append(data.Bridges, BridgeDataSourceModel{
			Name:      types.StringValue(bridge.Name),
			DHCP:      types.BoolValue(bridge.DHCP),
			Interface: types.StringValue(bridge.Interface),
		})
	}

	for _, bond := range result.NetworkProfile.Bonds {
		data.Bonds = append(data.Bonds, BondDataSourceModel{
			Name:       types.StringValue(bond.Name),
			Interfaces: types.StringValue(bond.Interfaces),
			Kind:       types.StringValue(bond.Kind),
			DHCP:       types.BoolValue(bond.DHCP),
		})
	}

	data.ID = types.Int64Value(result.NetworkProfile.ID)
	data.Name = types.StringValue(result.NetworkProfile.Name)
	data.Status = types.StringValue(result.NetworkProfile.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

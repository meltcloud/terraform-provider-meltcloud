package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NestInstanceDataSource{}

func NewNestInstanceDataSource() datasource.DataSource {
	return &NestInstanceDataSource{}
}

type NestInstanceDataSource struct {
	client *client.Client
}

type NestInstanceDataSourceModel struct {
	ID               types.Int64                       `tfsdk:"id"`
	NestID           types.Int64                       `tfsdk:"nest_id"`
	Name             types.String                      `tfsdk:"name"`
	ZoneID           types.Int64                       `tfsdk:"zone_id"`
	NetworkProfileID types.Int64                       `tfsdk:"network_profile_id"`
	IPAddresses      []NestInstanceIPAddressModel      `tfsdk:"ip_addresses"`
	BootDisk         NestInstanceBootDiskModel         `tfsdk:"boot_disk"`
	EtcdDisk         NestInstanceEtcdDiskModel         `tfsdk:"etcd_disk"`
	BGPLocalASN      types.Int64                       `tfsdk:"bgp_local_asn"`
	BGPPeers         []NestInstanceDataSourcePeerModel `tfsdk:"bgp_peers"`
	Status           types.String                      `tfsdk:"status"`
	Architecture     types.String                      `tfsdk:"architecture"`
	Version          types.String                      `tfsdk:"version"`
}

type NestInstanceIPAddressModel struct {
	IPPoolID types.Int64  `tfsdk:"ip_pool_id"`
	Address  types.String `tfsdk:"address"`
	Static   types.Bool   `tfsdk:"static"`
}

type NestInstanceDataSourcePeerModel struct {
	IP       types.String `tfsdk:"ip"`
	ASN      types.Int64  `tfsdk:"asn"`
	Multihop types.Bool   `tfsdk:"multihop"`
	BFD      types.Bool   `tfsdk:"bfd"`
}

func (d *NestInstanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_instance"
}

func (d *NestInstanceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	computedInt64 := func(description string) schema.Int64Attribute {
		return schema.Int64Attribute{Computed: true, MarkdownDescription: description}
	}
	computedBool := func(description string) schema.BoolAttribute {
		return schema.BoolAttribute{Computed: true, MarkdownDescription: description}
	}
	computedStr := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, MarkdownDescription: description}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A host added to the Nest as [Nest Instance](https://docs.meltcloud.io/tasks/nest/create-instance). Needs the API Key of the admin Organization.",
		Attributes: withLookupAttributes(map[string]schema.Attribute{
			"nest_id":            computedInt64("ID of the Nest the instance belongs to"),
			"zone_id":            computedInt64("ID of the [Zone](https://docs.meltcloud.io/concepts/regions-and-zones) the instance is in"),
			"network_profile_id": computedInt64("ID of the [Network Profile](https://docs.meltcloud.io/concepts/networking/network-profiles) the instance uses"),
			"ip_addresses": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The addresses of the instance, one per IP Pool of its Network Profile",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"ip_pool_id": computedInt64("ID of the IP Pool"),
					"address":    computedStr("The address"),
					"static":     computedBool("Whether it is a [static IP](https://docs.meltcloud.io/concepts/networking/subnet-addressing#static-ip-addresses) rather than one the pool allocated"),
				}},
			},
			"boot_disk": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The boot disk of the instance",
				Attributes: map[string]schema.Attribute{
					"device":        computedStr("Device the OS is installed to"),
					"mirror":        computedBool("Whether the boot disk is mirrored"),
					"mirror_device": computedStr("Device the boot disk is mirrored to"),
				},
			},
			"etcd_disk": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The disk etcd is kept on",
				Attributes: map[string]schema.Attribute{
					"enabled":       computedBool("Whether etcd is kept on a disk of its own"),
					"device":        computedStr("Device etcd is kept on"),
					"mirror":        computedBool("Whether the etcd disk is mirrored"),
					"mirror_device": computedStr("Device the etcd disk is mirrored to"),
				},
			},
			"bgp_local_asn": computedInt64("BGP ASN of the instance"),
			"bgp_peers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The BGP peers of the instance, without their passwords",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"ip":       computedStr("IP address of the peer"),
					"asn":      computedInt64("ASN of the peer"),
					"multihop": computedBool("Whether the peer is more than one hop away"),
					"bfd":      computedBool("Whether BFD is on"),
				}},
			},
			"status":       computedStr("Status of the instance: `installing` until it has joined the Nest, then `ready`. `boot_failed` if it reported a failed boot; see [Troubleshoot](https://docs.meltcloud.io/tasks/nest/troubleshoot#joining). A reset instance is `removed` until it installs again"),
			"architecture": computedStr("CPU architecture the instance reported once it booted"),
			"version":      computedStr("Nest version the instance runs"),
		}, "Internal ID of the Nest Instance on meltcloud", "Host name of the Nest Instance, unique within the Nest"),
	}
}

func (d *NestInstanceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *NestInstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NestInstanceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *client.NestInstanceResult
	var err *client.Error
	if data.ID.IsNull() {
		result, err = d.client.Nest().GetInstanceByName(ctx, data.Name.ValueString())
	} else {
		result, err = d.client.Nest().GetInstance(ctx, data.ID.ValueInt64())
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest Instance, got error: %s", err))
		return
	}

	instance := result.NestInstance
	data = NestInstanceDataSourceModel{
		ID:               types.Int64Value(instance.ID),
		NestID:           types.Int64Value(instance.NestID),
		Name:             types.StringValue(instance.Name),
		ZoneID:           types.Int64Value(instance.ZoneID),
		NetworkProfileID: types.Int64Value(instance.NetworkProfileID),
		IPAddresses:      []NestInstanceIPAddressModel{},
		BootDisk: NestInstanceBootDiskModel{
			Device:       types.StringValue(instance.InstallDiskDevice),
			Mirror:       types.BoolValue(instance.InstallDiskMirror),
			MirrorDevice: emptyAsNull(instance.InstallDiskMirrorDevice),
		},
		EtcdDisk: NestInstanceEtcdDiskModel{
			Enabled:      types.BoolValue(instance.EtcdDisk),
			Device:       emptyAsNull(instance.EtcdDiskDevice),
			Mirror:       types.BoolValue(instance.EtcdDiskMirror),
			MirrorDevice: emptyAsNull(instance.EtcdDiskMirrorDevice),
		},
		BGPLocalASN:  types.Int64PointerValue(instance.BGPLocalASN),
		BGPPeers:     []NestInstanceDataSourcePeerModel{},
		Status:       types.StringValue(instance.Status),
		Architecture: types.StringValue(instance.Architecture),
		Version:      types.StringValue(instance.Version),
	}
	for _, address := range instance.IPAddresses {
		data.IPAddresses = append(data.IPAddresses, NestInstanceIPAddressModel{
			IPPoolID: types.Int64Value(address.IPPoolID),
			Address:  types.StringValue(address.Address),
			Static:   types.BoolValue(address.Source == "operator"),
		})
	}
	for _, peer := range instance.BGPPeers {
		data.BGPPeers = append(data.BGPPeers, NestInstanceDataSourcePeerModel{
			IP:       types.StringValue(peer.IP),
			ASN:      types.Int64Value(peer.ASN),
			Multihop: types.BoolValue(peer.Multihop),
			BFD:      types.BoolValue(peer.BFD),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func configureClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *client.Client {
	if req.ProviderData == nil {
		return nil
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	return client
}

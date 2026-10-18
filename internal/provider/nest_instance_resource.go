package provider

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &NestInstanceResource{}
	_ resource.ResourceWithImportState = &NestInstanceResource{}
)

func NewNestInstanceResource() resource.Resource {
	return &NestInstanceResource{}
}

type NestInstanceResource struct {
	client *client.Client
}

type NestInstanceResourceModel struct {
	ID                      types.Int64           `tfsdk:"id"`
	NestID                  types.Int64           `tfsdk:"nest_id"`
	Name                    types.String          `tfsdk:"name"`
	ZoneID                  types.Int64           `tfsdk:"zone_id"`
	NetworkProfileID        types.Int64           `tfsdk:"network_profile_id"`
	IPAddresses             types.Set             `tfsdk:"ip_addresses"`
	InstallDiskDevice       types.String          `tfsdk:"install_disk_device"`
	InstallDiskMirror       types.Bool            `tfsdk:"install_disk_mirror"`
	InstallDiskMirrorDevice types.String          `tfsdk:"install_disk_mirror_device"`
	EtcdDisk                types.Bool            `tfsdk:"etcd_disk"`
	EtcdDiskDevice          types.String          `tfsdk:"etcd_disk_device"`
	EtcdDiskMirror          types.Bool            `tfsdk:"etcd_disk_mirror"`
	EtcdDiskMirrorDevice    types.String          `tfsdk:"etcd_disk_mirror_device"`
	BGP                     *NestInstanceBGPModel `tfsdk:"bgp"`
	Status                  types.String          `tfsdk:"status"`
	Architecture            types.String          `tfsdk:"architecture"`
	Version                 types.String          `tfsdk:"version"`
}

type NestInstanceBGPModel struct {
	LocalASN types.Int64                `tfsdk:"local_asn"`
	Peers    []NestInstanceBGPPeerModel `tfsdk:"peers"`
}

type NestInstanceBGPPeerModel struct {
	IP       types.String          `tfsdk:"ip"`
	ASN      types.Int64           `tfsdk:"asn"`
	Password types.String          `tfsdk:"password"`
	Multihop types.Bool            `tfsdk:"multihop"`
	BFD      *NestInstanceBFDModel `tfsdk:"bfd"`
}

type NestInstanceBFDModel struct {
	ReceiveInterval  types.Int64 `tfsdk:"receive_interval"`
	TransmitInterval types.Int64 `tfsdk:"transmit_interval"`
	DetectMultiplier types.Int64 `tfsdk:"detect_multiplier"`
}

func (r *NestInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_instance"
}

const nestInstanceDesc = "A [Nest Instance](https://docs.meltcloud.io/concepts/nest) is one machine of the Nest. " +
	"It installs and joins the Nest once its machine boots from a `meltcloud_nest_instance_image`; `status` turns from `installing` to `ready` then.\n\n" +
	"Needs the API Key of the admin Organization. A Nest has at most seven instances.\n\n" +
	"~> Only an instance that has not joined the Nest can be destroyed. To take out one that has, reset it in the admin area first; it is `removed` then.\n\n" +
	"~> After an instance joins, rebalance a highly available Foundry in the admin area."

func immutableString(description string, required bool) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Required:            required,
		Optional:            !required,
		Computed:            !required,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
	}
}

func immutableBool(description string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace(), boolplanmodifier.UseStateForUnknown()},
	}
}

func immutableInt64(description string, required bool) schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: description,
		Required:            required,
		Optional:            !required,
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
	}
}

func computedString(description string) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: description, Computed: true}
}

// The instance has no update, so any change to BGP replaces it. A password is the exception while the
// state does not know it, as after an import.
func bgpRequiresReplace(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		resp.RequiresReplace = req.StateValue.IsNull() != req.PlanValue.IsNull()
		return
	}

	var state, plan NestInstanceBGPModel
	if req.StateValue.As(ctx, &state, basetypes.ObjectAsOptions{}).HasError() ||
		req.PlanValue.As(ctx, &plan, basetypes.ObjectAsOptions{}).HasError() {
		resp.RequiresReplace = true
		return
	}
	resp.RequiresReplace = bgpChanged(state, plan)
}

func bgpChanged(state, plan NestInstanceBGPModel) bool {
	if !state.LocalASN.Equal(plan.LocalASN) || len(state.Peers) != len(plan.Peers) {
		return true
	}

	statePeers := map[string]NestInstanceBGPPeerModel{}
	for _, peer := range state.Peers {
		statePeers[peer.IP.ValueString()] = peer
	}
	for _, p := range plan.Peers {
		s, ok := statePeers[p.IP.ValueString()]
		if !ok || p.IP.IsUnknown() || !s.ASN.Equal(p.ASN) || !s.Multihop.Equal(p.Multihop) || bfdChanged(s.BFD, p.BFD) {
			return true
		}
		if !s.Password.IsNull() && !s.Password.Equal(p.Password) {
			return true
		}
	}
	return false
}

func bfdChanged(state, plan *NestInstanceBFDModel) bool {
	if state == nil || plan == nil {
		return (state == nil) != (plan == nil)
	}
	return !state.ReceiveInterval.Equal(plan.ReceiveInterval) ||
		!state.TransmitInterval.Equal(plan.TransmitInterval) ||
		!state.DetectMultiplier.Equal(plan.DetectMultiplier)
}

func optionalInt64(description string) schema.Int64Attribute {
	return schema.Int64Attribute{MarkdownDescription: description, Optional: true}
}

func bgpAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "BGP peering of the instance. Needed when the Nest uses BGP. Changing it replaces the instance",
		Optional:            true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplaceIf(bgpRequiresReplace,
			"Changing BGP replaces the Nest Instance", "Changing BGP replaces the Nest Instance")},
		Attributes: map[string]schema.Attribute{
			"local_asn": schema.Int64Attribute{MarkdownDescription: "ASN of the instance", Required: true},
			"peers": schema.ListNestedAttribute{
				MarkdownDescription: "The BGP peers of the instance, each with its own address",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip":  schema.StringAttribute{MarkdownDescription: "IP address of the peer", Required: true},
						"asn": schema.Int64Attribute{MarkdownDescription: "ASN of the peer", Required: true},
						"password": schema.StringAttribute{
							MarkdownDescription: "Password of the BGP session. It is never read back, so on an imported instance set it to the one in use",
							Optional:            true,
							Sensitive:           true,
						},
						"multihop": schema.BoolAttribute{
							MarkdownDescription: "Enable if the peer is not directly connected (more than one hop away). Defaults to `false`",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"bfd": schema.SingleNestedAttribute{
							MarkdownDescription: "Enables BFD on the session. Timings left out keep the defaults",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"receive_interval":  optionalInt64("Receive interval in milliseconds"),
								"transmit_interval": optionalInt64("Transmit interval in milliseconds"),
								"detect_multiplier": optionalInt64("Detect multiplier"),
							},
						},
					},
				},
			},
		},
	}
}

func (r *NestInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: nestInstanceDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Internal ID of the Nest Instance on meltcloud",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"nest_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "ID of the Nest the instance belongs to",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name":                       immutableString("Name of the Nest Instance, unique within the Nest", true),
			"zone_id":                    immutableInt64("ID of the [Zone](https://docs.meltcloud.io/concepts/regions-and-zones) the instance is in. Sets the Kubernetes labels `topology.kubernetes.io/region` and `topology.kubernetes.io/zone` on its node", true),
			"network_profile_id":         immutableInt64("ID of the [Network Profile](https://docs.meltcloud.io/concepts/networking/network-profiles) the instance uses", true),
			"install_disk_device":        immutableString("Device the instance installs to, for example `/dev/sda`. Leave it out to let the installer pick", false),
			"install_disk_mirror":        immutableBool("Mirror the install disk"),
			"install_disk_mirror_device": immutableString("Device the install disk is mirrored to", false),
			"etcd_disk":                  immutableBool("Keep etcd on a disk of its own"),
			"etcd_disk_device":           immutableString("Device etcd is kept on", false),
			"etcd_disk_mirror":           immutableBool("Mirror the etcd disk"),
			"etcd_disk_mirror_device":    immutableString("Device the etcd disk is mirrored to", false),
			"status":                     computedString("Status of the instance: `installing` until it has joined the Nest, then `ready`. A reset instance is `removed` until it installs again"),
			"architecture":               computedString("CPU architecture the instance reported once it booted"),
			"version":                    computedString("Nest version the instance runs"),
			"ip_addresses": schema.SetNestedAttribute{
				MarkdownDescription: "[Static IPs](https://docs.meltcloud.io/concepts/networking/subnet-addressing#static-ip-addresses) in the IP Pools of the Network Profile, one per pool. The instance needs one on the primary host network. In the other pools, it gets an address the pool allocates",
				Required:            true,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"ip_pool_id": schema.Int64Attribute{MarkdownDescription: "ID of the IP Pool the address is in", Required: true},
					"address":    schema.StringAttribute{MarkdownDescription: "The address the instance gets in that IP Pool", Required: true},
				}},
			},
			"bgp": bgpAttribute(),
		},
	}
}

func (r *NestInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NestInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NestInstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ipAddresses []IPAddressResourceModel
	resp.Diagnostics.Append(data.IPAddresses.ElementsAs(ctx, &ipAddresses, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.NestInstanceCreateInput{
		Name:                    data.Name.ValueString(),
		ZoneID:                  data.ZoneID.ValueInt64(),
		NetworkProfileID:        data.NetworkProfileID.ValueInt64(),
		IPAddresses:             ipAddressInput(ipAddresses),
		InstallDiskDevice:       knownString(data.InstallDiskDevice),
		InstallDiskMirror:       knownBool(data.InstallDiskMirror),
		InstallDiskMirrorDevice: knownString(data.InstallDiskMirrorDevice),
		EtcdDisk:                knownBool(data.EtcdDisk),
		EtcdDiskDevice:          knownString(data.EtcdDiskDevice),
		EtcdDiskMirror:          knownBool(data.EtcdDiskMirror),
		EtcdDiskMirrorDevice:    knownString(data.EtcdDiskMirrorDevice),
	}
	if data.BGP != nil {
		input.BGPLocalASN = data.BGP.LocalASN.ValueInt64Pointer()
		input.BGPPeers = bgpPeersInput(data.BGP.Peers)
	}

	result, err := r.client.Nest().CreateInstance(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Nest Instance, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(setNestInstanceValues(ctx, result.NestInstance, &data, data.BGP)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NestInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NestInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Nest().GetInstance(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest Instance, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(setNestInstanceValues(ctx, result.NestInstance, &data, data.BGP)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Every other change replaces the instance, so an update only sets BGP passwords the state did not know
// or the order of the peers.
func (r *NestInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NestInstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.BGP = plan.BGP
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NestInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NestInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Nest().DeleteInstance(ctx, data.ID.ValueInt64())
	if err != nil && err.HTTPStatusCode != 404 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Nest Instance, got error: %s", err))
	}
}

var nestInstanceImportIDPattern = regexp.MustCompile(`nest_instances/(\d+)`)

func (r *NestInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := nestInstanceImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", nestInstanceImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func bgpPeersInput(peers []NestInstanceBGPPeerModel) []client.BGPPeer {
	input := []client.BGPPeer{}
	for _, peer := range peers {
		p := client.BGPPeer{
			IP:       peer.IP.ValueString(),
			ASN:      peer.ASN.ValueInt64(),
			Password: peer.Password.ValueStringPointer(),
			Multihop: peer.Multihop.ValueBool(),
			BFD:      peer.BFD != nil,
		}
		if peer.BFD != nil {
			p.BFDReceiveInterval = peer.BFD.ReceiveInterval.ValueInt64Pointer()
			p.BFDTransmitInterval = peer.BFD.TransmitInterval.ValueInt64Pointer()
			p.BFDDetectMultiplier = peer.BFD.DetectMultiplier.ValueInt64Pointer()
		}
		input = append(input, p)
	}
	return input
}

// The API never returns the passwords, so they come from known, matched to the peers by address.
func bgpValues(instance *client.NestInstance, known *NestInstanceBGPModel) *NestInstanceBGPModel {
	if instance.BGPLocalASN == nil && len(instance.BGPPeers) == 0 {
		return nil
	}

	passwords := map[string]types.String{}
	if known != nil {
		for _, peer := range known.Peers {
			passwords[peer.IP.ValueString()] = peer.Password
		}
	}

	bgp := &NestInstanceBGPModel{LocalASN: types.Int64PointerValue(instance.BGPLocalASN), Peers: []NestInstanceBGPPeerModel{}}
	for _, peer := range instance.BGPPeers {
		password, ok := passwords[peer.IP]
		if !ok {
			password = types.StringNull()
		}
		model := NestInstanceBGPPeerModel{
			IP:       types.StringValue(peer.IP),
			ASN:      types.Int64Value(peer.ASN),
			Password: password,
			Multihop: types.BoolValue(peer.Multihop),
		}
		if peer.BFD {
			model.BFD = &NestInstanceBFDModel{
				ReceiveInterval:  types.Int64PointerValue(peer.BFDReceiveInterval),
				TransmitInterval: types.Int64PointerValue(peer.BFDTransmitInterval),
				DetectMultiplier: types.Int64PointerValue(peer.BFDDetectMultiplier),
			}
		}
		bgp.Peers = append(bgp.Peers, model)
	}

	// Known peers keep the order of known, so a reordered config does not show up again.
	order := map[string]int{}
	if known != nil {
		for i, peer := range known.Peers {
			order[peer.IP.ValueString()] = i
		}
	}
	sort.SliceStable(bgp.Peers, func(i, j int) bool {
		a, aKnown := order[bgp.Peers[i].IP.ValueString()]
		b, bKnown := order[bgp.Peers[j].IP.ValueString()]
		return aKnown && (!bKnown || a < b)
	})
	return bgp
}

var ipAddressObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{"ip_pool_id": types.Int64Type, "address": types.StringType}}

// Only the static IPs: the pools hand out the others.
func staticIPAddresses(ctx context.Context, addresses []client.IPAddress) (types.Set, diag.Diagnostics) {
	static := []IPAddressResourceModel{}
	for _, address := range addresses {
		if address.Source != "operator" {
			continue
		}
		static = append(static, IPAddressResourceModel{IPPoolID: types.Int64Value(address.IPPoolID), Address: types.StringValue(address.Address)})
	}
	return types.SetValueFrom(ctx, ipAddressObjectType, static)
}

func setNestInstanceValues(ctx context.Context, instance *client.NestInstance, data *NestInstanceResourceModel, knownBGP *NestInstanceBGPModel) diag.Diagnostics {
	data.ID = types.Int64Value(instance.ID)
	data.NestID = types.Int64Value(instance.NestID)
	data.Name = types.StringValue(instance.Name)
	data.ZoneID = types.Int64Value(instance.ZoneID)
	data.NetworkProfileID = types.Int64Value(instance.NetworkProfileID)
	data.InstallDiskDevice = types.StringValue(instance.InstallDiskDevice)
	data.InstallDiskMirror = types.BoolValue(instance.InstallDiskMirror)
	data.InstallDiskMirrorDevice = types.StringValue(instance.InstallDiskMirrorDevice)
	data.EtcdDisk = types.BoolValue(instance.EtcdDisk)
	data.EtcdDiskDevice = types.StringValue(instance.EtcdDiskDevice)
	data.EtcdDiskMirror = types.BoolValue(instance.EtcdDiskMirror)
	data.EtcdDiskMirrorDevice = types.StringValue(instance.EtcdDiskMirrorDevice)
	data.BGP = bgpValues(instance, knownBGP)
	data.Status = types.StringValue(instance.Status)
	data.Architecture = types.StringValue(instance.Architecture)
	data.Version = types.StringValue(instance.Version)

	var diags diag.Diagnostics
	data.IPAddresses, diags = staticIPAddresses(ctx, instance.IPAddresses)
	return diags
}

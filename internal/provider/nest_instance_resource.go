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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	ID               types.Int64                `tfsdk:"id"`
	NestID           types.Int64                `tfsdk:"nest_id"`
	Name             types.String               `tfsdk:"name"`
	ZoneID           types.Int64                `tfsdk:"zone_id"`
	NetworkProfileID types.Int64                `tfsdk:"network_profile_id"`
	IPAddresses      types.Set                  `tfsdk:"ip_addresses"`
	BootDisk         *NestInstanceBootDiskModel `tfsdk:"boot_disk"`
	EtcdDisk         types.Object               `tfsdk:"etcd_disk"`
	BGP              *NestInstanceBGPModel      `tfsdk:"bgp"`
	Status           types.String               `tfsdk:"status"`
	Architecture     types.String               `tfsdk:"architecture"`
	Version          types.String               `tfsdk:"version"`
}

type NestInstanceBootDiskModel struct {
	Device       types.String `tfsdk:"device"`
	Mirror       types.Bool   `tfsdk:"mirror"`
	MirrorDevice types.String `tfsdk:"mirror_device"`
}

type NestInstanceEtcdDiskModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	Device       types.String `tfsdk:"device"`
	Mirror       types.Bool   `tfsdk:"mirror"`
	MirrorDevice types.String `tfsdk:"mirror_device"`
}

var bfdObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"enabled": types.BoolType, "receive_interval": types.Int64Type, "transmit_interval": types.Int64Type, "detect_multiplier": types.Int64Type,
}}

// The foundry drops the timings of a peer without BFD, so they would show up as a change on every plan.
type bfdTimingsNeedEnabled struct{}

func (bfdTimingsNeedEnabled) Description(ctx context.Context) string {
	return "BFD timings need enabled = true"
}

func (v bfdTimingsNeedEnabled) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (bfdTimingsNeedEnabled) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var bfd NestInstanceBFDModel
	if req.ConfigValue.As(ctx, &bfd, basetypes.ObjectAsOptions{}).HasError() || bfd.Enabled.IsUnknown() || bfd.Enabled.ValueBool() {
		return
	}
	if !bfd.ReceiveInterval.IsNull() || !bfd.TransmitInterval.IsNull() || !bfd.DetectMultiplier.IsNull() {
		resp.Diagnostics.AddAttributeError(req.Path, "BFD Timings Without BFD", "Set enabled = true to use the BFD timings, or leave them out.")
	}
}

var etcdDiskObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"enabled": types.BoolType, "device": types.StringType, "mirror": types.BoolType, "mirror_device": types.StringType,
}}

type NestInstanceBGPModel struct {
	LocalASN types.Int64                `tfsdk:"local_asn"`
	Peers    []NestInstanceBGPPeerModel `tfsdk:"peers"`
}

type NestInstanceBGPPeerModel struct {
	IP                types.String          `tfsdk:"ip"`
	ASN               types.Int64           `tfsdk:"asn"`
	PasswordWO        types.String          `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64           `tfsdk:"password_wo_version"`
	Multihop          types.Bool            `tfsdk:"multihop"`
	BFD               *NestInstanceBFDModel `tfsdk:"bfd"`
}

type NestInstanceBFDModel struct {
	Enabled          types.Bool  `tfsdk:"enabled"`
	ReceiveInterval  types.Int64 `tfsdk:"receive_interval"`
	TransmitInterval types.Int64 `tfsdk:"transmit_interval"`
	DetectMultiplier types.Int64 `tfsdk:"detect_multiplier"`
}

func (r *NestInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_instance"
}

const nestInstanceDesc = "To add a host as [Nest Instance](https://docs.meltcloud.io/tasks/nest/create-instance), create it, then boot the host from a `meltcloud_nest_instance_image`. " +
	"`status` turns from `installing` to `ready` once it has joined the Nest.\n\n" +
	"Needs the API Key of the admin Organization. A Nest has at most seven instances.\n\n" +
	"The API can't update an instance, so any change replaces it.\n\n" +
	"~> Only an instance that has not joined the Nest can be destroyed. To take out one that has, [reset](https://docs.meltcloud.io/tasks/nest/reset-instance) it in the admin area first; it is `removed` then.\n\n" +
	"~> After an instance joins, [enable high availability](https://docs.meltcloud.io/tasks/foundry/enable-high-availability) on a single-instance Foundry once three instances are `ready`, or [rebalance](https://docs.meltcloud.io/tasks/foundry/rebalance) a highly available one, in the admin area."

func immutableString(description string, required bool) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Required:            required,
		Optional:            !required,
		Computed:            !required,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
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

// The instance has no update, so any change to BGP replaces it. A password version is the exception while
// the state has none, as after an import.
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
		if !s.PasswordWOVersion.IsNull() && !s.PasswordWOVersion.Equal(p.PasswordWOVersion) {
			return true
		}
	}
	return false
}

func bfdChanged(state, plan *NestInstanceBFDModel) bool {
	if state == nil || plan == nil {
		return (state == nil) != (plan == nil)
	}
	return !state.Enabled.Equal(plan.Enabled) ||
		!state.ReceiveInterval.Equal(plan.ReceiveInterval) ||
		!state.TransmitInterval.Equal(plan.TransmitInterval) ||
		!state.DetectMultiplier.Equal(plan.DetectMultiplier)
}

const diskDeviceHint = "Use a `/dev/disk/by-path/` path rather than a kernel name such as `/dev/sda`, which can change between boots"

// disk is "boot" or "etcd".
func mirrorAttributes(disk string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"mirror": schema.BoolAttribute{
			MarkdownDescription: "Mirror the " + disk + " disk onto a second device. Defaults to `false`",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"mirror_device": schema.StringAttribute{MarkdownDescription: "Block device to use as the " + disk + " mirror disk. Only with `mirror`", Optional: true},
	}
}

func bootDiskAttribute() schema.SingleNestedAttribute {
	attributes := mirrorAttributes("boot")
	attributes["device"] = schema.StringAttribute{
		MarkdownDescription: "Block device to install the OS to, for example `/dev/disk/by-path/pci-0000:00:17.0-ata-1`. " + diskDeviceHint,
		Required:            true,
	}
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The boot disk of the instance",
		Required:            true,
		PlanModifiers:       []planmodifier.Object{objectplanmodifier.RequiresReplace()},
		Attributes:          attributes,
	}
}

func etcdDiskAttribute() schema.SingleNestedAttribute {
	attributes := mirrorAttributes("etcd")
	attributes["enabled"] = schema.BoolAttribute{MarkdownDescription: "Keep etcd on a disk of its own", Required: true}
	attributes["device"] = schema.StringAttribute{MarkdownDescription: "Block device to use for etcd. Only with `enabled`. " + diskDeviceHint, Optional: true}
	return schema.SingleNestedAttribute{
		MarkdownDescription: "A dedicated disk for etcd. Leave it out to keep etcd on the boot disk",
		Optional:            true,
		Computed:            true,
		Default: objectdefault.StaticValue(types.ObjectValueMust(etcdDiskObjectType.AttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(false), "device": types.StringNull(), "mirror": types.BoolValue(false), "mirror_device": types.StringNull(),
		})),
		PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
		Attributes:    attributes,
	}
}

func optionalInt64(description string) schema.Int64Attribute {
	return schema.Int64Attribute{MarkdownDescription: description, Optional: true}
}

func bgpAttribute() schema.SingleNestedAttribute {
	peer := map[string]schema.Attribute{
		"ip":  schema.StringAttribute{MarkdownDescription: "IP address of the peer", Required: true},
		"asn": schema.Int64Attribute{MarkdownDescription: "ASN of the peer", Required: true},
		"multihop": schema.BoolAttribute{
			MarkdownDescription: "Enable if the peer is not directly connected (more than one hop away). Defaults to `false`",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"bfd": schema.SingleNestedAttribute{
			MarkdownDescription: "BFD on the session. Leave it out to keep BFD off",
			Optional:            true,
			Computed:            true,
			Default: objectdefault.StaticValue(types.ObjectValueMust(bfdObjectType.AttrTypes, map[string]attr.Value{
				"enabled": types.BoolValue(false), "receive_interval": types.Int64Null(),
				"transmit_interval": types.Int64Null(), "detect_multiplier": types.Int64Null(),
			})),
			Validators: []validator.Object{bfdTimingsNeedEnabled{}},
			Attributes: map[string]schema.Attribute{
				"enabled":           schema.BoolAttribute{MarkdownDescription: "Enable BFD on the session", Required: true},
				"receive_interval":  optionalInt64("Receive interval in milliseconds. Only with `enabled`; left out, it keeps the default"),
				"transmit_interval": optionalInt64("Transmit interval in milliseconds. Only with `enabled`; left out, it keeps the default"),
				"detect_multiplier": optionalInt64("Detect multiplier. Only with `enabled`; left out, it keeps the default"),
			},
		},
	}
	writeOnlySecret(peer, "password_wo", "Password of the BGP session",
		"On an imported instance, set it to the password in use: the first version is recorded without a replace")
	return schema.SingleNestedAttribute{
		MarkdownDescription: "BGP peering of the instance. Needed when the Nest uses BGP",
		Optional:            true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplaceIf(bgpRequiresReplace,
			"Changing BGP replaces the Nest Instance", "Changing BGP replaces the Nest Instance")},
		Attributes: map[string]schema.Attribute{
			"local_asn": schema.Int64Attribute{MarkdownDescription: "ASN of the instance", Required: true},
			"peers": schema.ListNestedAttribute{
				MarkdownDescription: "The BGP peers of the instance, each with its own address",
				Required:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: peer},
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
			"name":               immutableString("Host name of the Nest Instance, unique within the Nest", true),
			"zone_id":            immutableInt64("ID of the [Zone](https://docs.meltcloud.io/concepts/regions-and-zones) the instance is in. Sets the Kubernetes labels `topology.kubernetes.io/region` and `topology.kubernetes.io/zone` on its node", true),
			"network_profile_id": immutableInt64("ID of the [Network Profile](https://docs.meltcloud.io/concepts/networking/network-profiles) the instance uses", true),
			"boot_disk":          bootDiskAttribute(),
			"etcd_disk":          etcdDiskAttribute(),
			"status":             computedString("Status of the instance: `installing` until it has joined the Nest, then `ready`. `boot_failed` if it reported a failed boot; see [Troubleshoot](https://docs.meltcloud.io/tasks/nest/troubleshoot#joining). A reset instance is `removed` until it installs again"),
			"architecture":       computedString("CPU architecture the instance reported"),
			"version":            computedString("Nest version the instance runs"),
			"ip_addresses": schema.SetNestedAttribute{
				MarkdownDescription: "[Static IPs](https://docs.meltcloud.io/concepts/networking/subnet-addressing#static-ip-addresses) in the IP Pools of the Network Profile, one per pool. The instance needs one on the primary host network. In the other pools, it gets an address the pool allocates",
				Required:            true,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"ip_pool_id": schema.Int64Attribute{MarkdownDescription: "ID of the IP Pool the address is in", Required: true},
					"address":    schema.StringAttribute{MarkdownDescription: "IP address of the instance in that IP Pool", Required: true},
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
	var data, config NestInstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
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
		InstallDiskDevice:       knownString(data.BootDisk.Device),
		InstallDiskMirror:       knownBool(data.BootDisk.Mirror),
		InstallDiskMirrorDevice: knownString(data.BootDisk.MirrorDevice),
	}
	if !data.EtcdDisk.IsNull() && !data.EtcdDisk.IsUnknown() {
		var etcd NestInstanceEtcdDiskModel
		resp.Diagnostics.Append(data.EtcdDisk.As(ctx, &etcd, basetypes.ObjectAsOptions{})...)
		input.EtcdDisk = knownBool(etcd.Enabled)
		input.EtcdDiskDevice = knownString(etcd.Device)
		input.EtcdDiskMirror = knownBool(etcd.Mirror)
		input.EtcdDiskMirrorDevice = knownString(etcd.MirrorDevice)
	}
	if data.BGP != nil {
		input.BGPLocalASN = data.BGP.LocalASN.ValueInt64Pointer()
		input.BGPPeers = bgpPeersInput(data.BGP.Peers, config.BGP.Peers)
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

// Every other change replaces the instance, so an update only records a first password version or the
// order of the peers.
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

// The passwords come from the config: being write-only, the plan has none.
func bgpPeersInput(peers, configPeers []NestInstanceBGPPeerModel) []client.BGPPeer {
	input := []client.BGPPeer{}
	for i, peer := range peers {
		p := client.BGPPeer{
			IP:       peer.IP.ValueString(),
			ASN:      peer.ASN.ValueInt64(),
			Password: knownString(configPeers[i].PasswordWO),
			Multihop: peer.Multihop.ValueBool(),
			BFD:      peer.BFD != nil && peer.BFD.Enabled.ValueBool(),
		}
		if p.BFD {
			p.BFDReceiveInterval = peer.BFD.ReceiveInterval.ValueInt64Pointer()
			p.BFDTransmitInterval = peer.BFD.TransmitInterval.ValueInt64Pointer()
			p.BFDDetectMultiplier = peer.BFD.DetectMultiplier.ValueInt64Pointer()
		}
		input = append(input, p)
	}
	return input
}

// The API never returns the passwords, so their versions come from known, matched to the peers by address.
func bgpValues(instance *client.NestInstance, known *NestInstanceBGPModel) *NestInstanceBGPModel {
	if instance.BGPLocalASN == nil && len(instance.BGPPeers) == 0 {
		return nil
	}

	versions := map[string]types.Int64{}
	if known != nil {
		for _, peer := range known.Peers {
			versions[peer.IP.ValueString()] = peer.PasswordWOVersion
		}
	}

	bgp := &NestInstanceBGPModel{LocalASN: types.Int64PointerValue(instance.BGPLocalASN), Peers: []NestInstanceBGPPeerModel{}}
	for _, peer := range instance.BGPPeers {
		version, ok := versions[peer.IP]
		if !ok {
			version = types.Int64Null()
		}
		model := NestInstanceBGPPeerModel{
			IP:                types.StringValue(peer.IP),
			ASN:               types.Int64Value(peer.ASN),
			PasswordWO:        types.StringNull(),
			PasswordWOVersion: version,
			Multihop:          types.BoolValue(peer.Multihop),
		}
		model.BFD = &NestInstanceBFDModel{
			Enabled:          types.BoolValue(peer.BFD),
			ReceiveInterval:  types.Int64Null(),
			TransmitInterval: types.Int64Null(),
			DetectMultiplier: types.Int64Null(),
		}
		if peer.BFD {
			model.BFD.ReceiveInterval = types.Int64PointerValue(peer.BFDReceiveInterval)
			model.BFD.TransmitInterval = types.Int64PointerValue(peer.BFDTransmitInterval)
			model.BFD.DetectMultiplier = types.Int64PointerValue(peer.BFDDetectMultiplier)
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
	data.BootDisk = &NestInstanceBootDiskModel{
		Device:       types.StringValue(instance.InstallDiskDevice),
		Mirror:       types.BoolValue(instance.InstallDiskMirror),
		MirrorDevice: emptyAsNull(instance.InstallDiskMirrorDevice),
	}
	data.BGP = bgpValues(instance, knownBGP)
	data.Status = types.StringValue(instance.Status)
	data.Architecture = types.StringValue(instance.Architecture)
	data.Version = types.StringValue(instance.Version)

	var diags, etcdDiags diag.Diagnostics
	data.EtcdDisk, etcdDiags = types.ObjectValueFrom(ctx, etcdDiskObjectType.AttrTypes, NestInstanceEtcdDiskModel{
		Enabled:      types.BoolValue(instance.EtcdDisk),
		Device:       emptyAsNull(instance.EtcdDiskDevice),
		Mirror:       types.BoolValue(instance.EtcdDiskMirror),
		MirrorDevice: emptyAsNull(instance.EtcdDiskMirrorDevice),
	})
	data.IPAddresses, diags = staticIPAddresses(ctx, instance.IPAddresses)
	return append(diags, etcdDiags...)
}

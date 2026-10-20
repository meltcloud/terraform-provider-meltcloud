package provider

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &MachineResource{}
	_ resource.ResourceWithImportState    = &MachineResource{}
	_ resource.ResourceWithValidateConfig = &MachineResource{}
)

func NewMachineResource() resource.Resource {
	return &MachineResource{}
}

// MachineResource defines the resource implementation.
type MachineResource struct {
	client *client.Client
}

// MachineResourceModel describes the resource data model.
type MachineResourceModel struct {
	ID                    types.Int64  `tfsdk:"id"`
	UUID                  types.String `tfsdk:"uuid"`
	Name                  types.String `tfsdk:"name"`
	MachinePoolID         types.Int64  `tfsdk:"machine_pool_id"`
	NetworkProfileID      types.Int64  `tfsdk:"network_profile_id"`
	DepotNetworkProfileID types.Int64  `tfsdk:"depot_network_profile_id"`
	ZoneID                types.Int64  `tfsdk:"zone_id"`
	Labels                types.Map    `tfsdk:"labels"`
	IPAddresses           types.Set    `tfsdk:"ip_addresses"`
}

type IPAddressResourceModel struct {
	IPPoolID types.Int64  `tfsdk:"ip_pool_id"`
	Address  types.String `tfsdk:"address"`
}

func (r *MachineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

const machineDesc string = "[Machines](https://docs.meltcloud.io/concepts/machines) are bare-metal or virtualized computers designated as worker nodes for the Kubernetes control planes provided by the [Nest](https://docs.meltcloud.io/concepts/nest) appliance."

func machineResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Machine in meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"uuid": schema.StringAttribute{
			MarkdownDescription: "UUID of the Machine",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Machine",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"machine_pool_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the associated machine pool",
			Optional:            true,
		},
		"network_profile_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Network Profile the Machine runs with",
			Required:            true,
		},
		"depot_network_profile_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Network Profile the Machine uses while it has no pool, and in recovery",
			Required:            true,
		},
		"zone_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the [Zone](https://docs.meltcloud.io/concepts/regions-and-zones) the Machine is in. Sets the Kubernetes labels `topology.kubernetes.io/region` and `topology.kubernetes.io/zone` on the Machine",
			Optional:            true,
		},
		"labels": schema.MapAttribute{
			MarkdownDescription: "Kubernetes labels to set on the Machine's node, for example `{ \"node-role.example.com/storage\" = \"true\" }`",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Map{
				mapvalidator.SizeAtLeast(1),
			},
		},
		"ip_addresses": schema.SetNestedAttribute{
			MarkdownDescription: "[Static IPs](https://docs.meltcloud.io/concepts/networking/subnet-addressing#static-ip-addresses) in the IP Pools the Machine's Network Profiles use, at most one per IP Pool. In a pool without one, the Machine gets an address the pool allocates",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: ipAddressResourceAttributes(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
	}
}

func ipAddressResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"ip_pool_id": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "ID of the IP Pool the address is in. The Machine has one address per IP Pool",
		},
		"address": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The address the Machine always gets in that IP Pool",
		},
	}
}

func (r *MachineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: machineDesc + "\n\n" +
			"This resource [pre-registers](https://docs.meltcloud.io/tasks/machines/add#pre-register) Machines for a later boot.\n\n" +
			"~> Be aware that changing the name will cause a new [Revision that will be applied immediately, causing a reboot of the Machine](https://docs.meltcloud.io/concepts/machines#revisions).",

		Attributes: machineResourceAttributes(),
	}
}

func (r *MachineResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() || data.IPAddresses.IsNull() || data.IPAddresses.IsUnknown() {
		return
	}

	pools := map[int64]bool{}
	for _, element := range data.IPAddresses.Elements() {
		var address IPAddressResourceModel
		object, ok := element.(types.Object)
		if !ok {
			continue
		}
		resp.Diagnostics.Append(object.As(ctx, &address, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() || address.IPPoolID.IsUnknown() {
			continue
		}

		if pools[address.IPPoolID.ValueInt64()] {
			resp.Diagnostics.AddAttributeError(
				path.Root("ip_addresses").AtSetValue(element),
				"Invalid Attribute Value",
				fmt.Sprintf("The Machine has one address per IP Pool, but IP Pool %d appears more than once.", address.IPPoolID.ValueInt64()),
			)
		}
		pools[address.IPPoolID.ValueInt64()] = true
	}
}

func (r *MachineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *MachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uuid, err := uuid.Parse(data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("UUID invalid: %s", err))
		return
	}

	labels := map[string]string{}
	resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ipAddresses []IPAddressResourceModel
	resp.Diagnostics.Append(data.IPAddresses.ElementsAs(ctx, &ipAddresses, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	machineCreateInput := &client.MachineCreateInput{
		UUID:                  uuid,
		Name:                  optionalString(data.Name),
		MachinePoolID:         data.MachinePoolID.ValueInt64(),
		NetworkProfileID:      data.NetworkProfileID.ValueInt64(),
		DepotNetworkProfileID: data.DepotNetworkProfileID.ValueInt64(),
		ZoneID:                data.ZoneID.ValueInt64Pointer(),
		Labels:                labelInput(labels),
		IPAddresses:           ipAddressInput(ipAddresses),
	}

	result, err2 := r.client.Machine().Create(ctx, machineCreateInput)
	if err2 != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create machine, got error: %s", err2))
		return
	}

	data.ID = types.Int64Value(result.Machine.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Machine().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine, got error: %s", err))
		return
	}

	data.UUID = types.StringValue(result.Machine.UUID.String())
	data.Name = types.StringPointerValue(result.Machine.Name)
	data.MachinePoolID = types.Int64PointerValue(result.Machine.MachinePoolID)
	data.NetworkProfileID = types.Int64Value(result.Machine.NetworkProfileID)
	data.DepotNetworkProfileID = types.Int64Value(result.Machine.DepotNetworkProfileID)
	data.ZoneID = types.Int64PointerValue(result.Machine.ZoneID)

	labels, diags := labelValues(ctx, result.Machine.Labels)
	resp.Diagnostics.Append(diags...)
	ipAddresses, diags := ipAddressValues(ctx, result.Machine.Revision)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Labels = labels
	data.IPAddresses = ipAddresses

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Like the other optional collections, no labels and no static IPs read as
// absent: the attributes refuse an empty map or set, so that is what was configured.
func labelValues(ctx context.Context, labels []client.Label) (types.Map, diag.Diagnostics) {
	if len(labels) == 0 {
		return types.MapNull(types.StringType), nil
	}

	values := map[string]string{}
	for _, label := range labels {
		values[label.Key] = label.Value
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}

func ipAddressValues(ctx context.Context, revision *client.MachineRevision) (types.Set, diag.Diagnostics) {
	var ipAddresses []IPAddressResourceModel
	if revision != nil {
		for _, address := range revision.IPAddresses {
			if address.Source != "operator" {
				continue
			}
			ipAddresses = append(ipAddresses, IPAddressResourceModel{
				IPPoolID: types.Int64Value(address.IPPoolID),
				Address:  types.StringValue(address.Address),
			})
		}
	}

	if len(ipAddresses) == 0 {
		return types.SetNull(ipAddressObjectType), nil
	}
	return types.SetValueFrom(ctx, ipAddressObjectType, ipAddresses)
}

func (r *MachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	labels := map[string]string{}
	resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ipAddresses []IPAddressResourceModel
	resp.Diagnostics.Append(data.IPAddresses.ElementsAs(ctx, &ipAddresses, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	machineUpdateInput := &client.MachineUpdateInput{
		Name:                  optionalString(data.Name),
		MachinePoolID:         data.MachinePoolID.ValueInt64(),
		NetworkProfileID:      data.NetworkProfileID.ValueInt64(),
		DepotNetworkProfileID: data.DepotNetworkProfileID.ValueInt64(),
		ZoneID:                data.ZoneID.ValueInt64Pointer(),
		Labels:                labelInput(labels),
		IPAddresses:           ipAddressInput(ipAddresses),
	}

	result, err := r.client.Machine().Update(ctx, data.ID.ValueInt64(), machineUpdateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine, got error: %s", err))
			return
		}

		_, err := r.client.Machine().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine, got error: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Machine().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete machine, got error: %s", err))
		return
	}
}

var machineImportIDPattern = regexp.MustCompile(`machines/(\d+)`)

func (r *MachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := machineImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", machineImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Sorted by key, so the Machine's labels don't change order between applies.
func labelInput(labels map[string]string) []client.Label {
	labelInput := []client.Label{}
	for _, key := range slices.Sorted(maps.Keys(labels)) {
		labelInput = append(labelInput, client.Label{
			Key:   key,
			Value: labels[key],
		})
	}
	return labelInput
}

// Only static IPs: the pools hand out the others.
func ipAddressInput(addresses []IPAddressResourceModel) []client.IPAddress {
	input := []client.IPAddress{}
	for _, address := range addresses {
		input = append(input, client.IPAddress{
			IPPoolID: address.IPPoolID.ValueInt64(),
			Address:  address.Address.ValueString(),
		})
	}
	return input
}

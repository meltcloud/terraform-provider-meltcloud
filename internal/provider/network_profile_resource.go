package provider

import (
	"context"

	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NetworkProfileResource{}
var _ resource.ResourceWithImportState = &NetworkProfileResource{}

func NewNetworkProfileResource() resource.Resource {
	return &NetworkProfileResource{}
}

// NetworkProfileResource defines the resource implementation.
type NetworkProfileResource struct {
	client *client.Client
}

// NetworkProfileResourceModel describes the resource data model.
type NetworkProfileResourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Uplinks types.List   `tfsdk:"uplink"`
}

type UplinkResourceModel struct {
	Name         types.String `tfsdk:"name"`
	Mode         types.String `tfsdk:"mode"`
	Identifier   types.String `tfsdk:"identifier"`
	Interfaces   types.List   `tfsdk:"interfaces"`
	LACP         types.Bool   `tfsdk:"lacp"`
	HostNetworks types.List   `tfsdk:"host_network"`
}

type HostNetworkResourceModel struct {
	SubnetID   types.Int64 `tfsdk:"subnet_id"`
	VLANTagged types.Bool  `tfsdk:"vlan_tagged"`
	Primary    types.Bool  `tfsdk:"primary"`
}

func (r *NetworkProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_profile"
}

const networkProfileDesc = "A [Network Profile](https://docs.meltcloud.io/concepts/networking/network-profiles) specifies the network configuration for [Machines](https://docs.meltcloud.io/concepts/machines) in a [Machine Pool](https://docs.meltcloud.io/tasks/machine-pools/create)."

func networkProfileResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the network profile on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the network profile",
			Required:            true,
		},
	}
}

func uplinkResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the uplink, at most 10 lowercase alphanumeric characters",
		},
		"mode": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "How many interfaces the uplink expects: `auto` for the machine's only interface, `single` for one named interface, `bond` for several bonded together",
		},
		"identifier": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "What the interfaces are matched against: `kernel_name` (the default) or `mac_address`",
		},
		"interfaces": schema.ListAttribute{
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The interfaces to match. Empty with mode `auto`, one with `single`, at least two with `bond`",
		},
		"lacp": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Whether the bond runs LACP. Only available with mode `bond`",
		},
	}
}

func hostNetworkResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"subnet_id": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "ID of the Subnet the machine is addressed on",
		},
		"vlan_tagged": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether the subnet's VLAN arrives tagged, which configures a VLAN subinterface. At most one untagged host network per uplink",
		},
		"primary": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether this host network supplies the default route, DNS and NTP. Exactly one across the profile",
		},
	}
}

func (r *NetworkProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: networkProfileDesc,

		Attributes: networkProfileResourceAttributes(),

		Blocks: map[string]schema.Block{
			"uplink": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: uplinkResourceAttributes(),
					Blocks: map[string]schema.Block{
						"host_network": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: hostNetworkResourceAttributes(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *NetworkProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NetworkProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var uplinks []UplinkResourceModel
	diags := data.Uplinks.ElementsAs(ctx, &uplinks, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	networkProfileCreateInput := &client.NetworkProfileCreateInput{
		Name:    data.Name.ValueString(),
		Uplinks: r.uplinksInput(ctx, uplinks),
	}

	result, err := r.client.NetworkProfile().Create(ctx, networkProfileCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create network profile, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.NetworkProfile.ID)

	uplinkValues, diags := uplinkModels(ctx, result.NetworkProfile.Uplinks)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uplinkList, diags := types.ListValueFrom(ctx, uplinkObjectType(), uplinkValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Uplinks = uplinkList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func uplinkObjectType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":         types.StringType,
		"mode":         types.StringType,
		"identifier":   types.StringType,
		"interfaces":   types.ListType{ElemType: types.StringType},
		"lacp":         types.BoolType,
		"host_network": types.ListType{ElemType: hostNetworkObjectType()},
	}}
}

func uplinkModels(ctx context.Context, uplinks []client.Uplink) ([]UplinkResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var models []UplinkResourceModel

	for _, uplink := range uplinks {
		interfaces, interfaceDiags := types.ListValueFrom(ctx, types.StringType, uplink.Interfaces)
		diags.Append(interfaceDiags...)

		hostNetworks, hostNetworkDiags := types.ListValueFrom(ctx, hostNetworkObjectType(), hostNetworkValues(uplink))
		diags.Append(hostNetworkDiags...)
		if diags.HasError() {
			return nil, diags
		}

		models = append(models, UplinkResourceModel{
			Name:         types.StringValue(uplink.Name),
			Mode:         types.StringValue(uplink.Mode),
			Identifier:   types.StringValue(uplink.Identifier),
			Interfaces:   interfaces,
			LACP:         types.BoolValue(uplink.LACP),
			HostNetworks: hostNetworks,
		})
	}

	return models, diags
}

func (r *NetworkProfileResource) uplinksInput(ctx context.Context, uplinks []UplinkResourceModel) []client.Uplink {
	var uplinksInput []client.Uplink
	for _, uplink := range uplinks {
		var interfaces []string
		uplink.Interfaces.ElementsAs(ctx, &interfaces, false)

		var hostNetworks []HostNetworkResourceModel
		uplink.HostNetworks.ElementsAs(ctx, &hostNetworks, false)

		var hostNetworksInput []client.HostNetwork
		for _, hostNetwork := range hostNetworks {
			hostNetworksInput = append(hostNetworksInput, client.HostNetwork{
				SubnetID:   hostNetwork.SubnetID.ValueInt64(),
				VLANTagged: hostNetwork.VLANTagged.ValueBool(),
				Primary:    hostNetwork.Primary.ValueBool(),
			})
		}

		uplinksInput = append(uplinksInput, client.Uplink{
			Name:         uplink.Name.ValueString(),
			Mode:         uplink.Mode.ValueString(),
			Identifier:   uplink.Identifier.ValueString(),
			Interfaces:   interfaces,
			LACP:         uplink.LACP.ValueBool(),
			HostNetworks: hostNetworksInput,
		})
	}
	return uplinksInput
}

func (r *NetworkProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.NetworkProfile().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network profile, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.NetworkProfile.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var uplinks []UplinkResourceModel
	for _, uplink := range result.NetworkProfile.Uplinks {
		interfacesList, diags := types.ListValueFrom(ctx, types.StringType, uplink.Interfaces)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		hostNetworksList, diags := types.ListValueFrom(ctx, hostNetworkObjectType(), hostNetworkValues(uplink))
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		uplinks = append(uplinks, UplinkResourceModel{
			Name:         types.StringValue(uplink.Name),
			Mode:         types.StringValue(uplink.Mode),
			Identifier:   types.StringValue(uplink.Identifier),
			Interfaces:   interfacesList,
			LACP:         types.BoolValue(uplink.LACP),
			HostNetworks: hostNetworksList,
		})
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uplink"), uplinks)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func hostNetworkObjectType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"subnet_id":   types.Int64Type,
		"vlan_tagged": types.BoolType,
		"primary":     types.BoolType,
	}}
}

func hostNetworkValues(uplink client.Uplink) []HostNetworkResourceModel {
	var hostNetworks []HostNetworkResourceModel
	for _, hostNetwork := range uplink.HostNetworks {
		hostNetworks = append(hostNetworks, HostNetworkResourceModel{
			SubnetID:   types.Int64Value(hostNetwork.SubnetID),
			VLANTagged: types.BoolValue(hostNetwork.VLANTagged),
			Primary:    types.BoolValue(hostNetwork.Primary),
		})
	}
	return hostNetworks
}

// A network profile is immutable in foundry, so every attribute replaces the
// resource and Update is never called.
func (r *NetworkProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Supported", "A network profile cannot be changed; it is replaced.")
}

func (r *NetworkProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.NetworkProfile().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete network profile, got error: %s", err))
		return
	}
}

var networkProfileImportIDPattern = regexp.MustCompile(`network_profiles/(\d+)`)

func (r *NetworkProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := networkProfileImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", networkProfileImportIDPattern.String()))
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

package provider

import (
	"context"
	"fmt"
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
	VLANs   types.List   `tfsdk:"vlan"`
	Bridges types.List   `tfsdk:"bridge"`
	Bonds   types.List   `tfsdk:"bond"`
}

type VLANResourceModel struct {
	VLAN      types.Int64  `tfsdk:"vlan"`
	Interface types.String `tfsdk:"interface"`
	DHCP      types.Bool   `tfsdk:"dhcp"`
}

type BridgeResourceModel struct {
	Name      types.String `tfsdk:"name"`
	Interface types.String `tfsdk:"interface"`
	DHCP      types.Bool   `tfsdk:"dhcp"`
}

type BondResourceModel struct {
	Name       types.String `tfsdk:"name"`
	Interfaces types.String `tfsdk:"interfaces"`
	Kind       types.String `tfsdk:"kind"`
	DHCP       types.Bool   `tfsdk:"dhcp"`
}

func (r *NetworkProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_profile"
}

const networkProfileDesc = "A Network Profile specifies the network configuration to be used for machines in a machine pool."

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

func vlanResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"vlan": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "Vlan number",
		},

		"interface": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Interface list (systemd network config format)",
		},

		"dhcp": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether to use DHCP",
		},
	}
}

func bridgeResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Bridge name",
		},
		"interface": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Interface name",
		},
		"dhcp": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether to use DHCP",
		},
	}
}

func bondResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Bond name",
		},
		"interfaces": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Interface list (systemd network config format)",
		},
		"kind": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Bonding mode",
		},
		"dhcp": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether to use DHCP",
		},
	}
}

func (r *NetworkProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: networkProfileDesc,

		Attributes: networkProfileResourceAttributes(),

		Blocks: map[string]schema.Block{
			"vlan": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: vlanResourceAttributes(),
				},
			},
			"bridge": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: bridgeResourceAttributes(),
				},
			},
			"bond": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: bondResourceAttributes(),
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

	var vlans []VLANResourceModel
	diags := data.VLANs.ElementsAs(ctx, &vlans, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	var bridges []BridgeResourceModel
	diags = data.Bridges.ElementsAs(ctx, &bridges, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	var bonds []BondResourceModel
	diags = data.Bonds.ElementsAs(ctx, &bonds, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	networkProfileCreateInput := &client.NetworkProfileCreateInput{
		Name:    data.Name.ValueString(),
		VLANs:   r.vlansInput(vlans),
		Bridges: r.bridgesInput(bridges),
		Bonds:   r.bondsInput(bonds),
	}

	result, err := r.client.NetworkProfile().Create(ctx, networkProfileCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create network profile, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.NetworkProfile.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkProfileResource) vlansInput(vlans []VLANResourceModel) []client.VLAN {
	var vlansInput []client.VLAN
	for _, networkConfiguration := range vlans {
		vlansInput = append(vlansInput, client.VLAN{
			VLAN:      networkConfiguration.VLAN.ValueInt64(),
			Interface: networkConfiguration.Interface.ValueString(),
			DHCP:      networkConfiguration.DHCP.ValueBool(),
		})
	}
	return vlansInput
}

func (r *NetworkProfileResource) bridgesInput(bridges []BridgeResourceModel) []client.Bridge {
	var bridgesInput []client.Bridge
	for _, networkConfiguration := range bridges {
		bridgesInput = append(bridgesInput, client.Bridge{
			Name:      networkConfiguration.Name.ValueString(),
			Interface: networkConfiguration.Interface.ValueString(),
			DHCP:      networkConfiguration.DHCP.ValueBool(),
		})
	}
	return bridgesInput
}

func (r *NetworkProfileResource) bondsInput(bonds []BondResourceModel) []client.Bond {
	var bondsInput []client.Bond
	for _, networkConfiguration := range bonds {
		bondsInput = append(bondsInput, client.Bond{
			Name:       networkConfiguration.Name.ValueString(),
			Interfaces: networkConfiguration.Interfaces.ValueString(),
			Kind:       networkConfiguration.Kind.ValueString(),
			DHCP:       networkConfiguration.DHCP.ValueBool(),
		})
	}
	return bondsInput
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

	var vlans []VLANResourceModel
	for _, vlan := range result.NetworkProfile.VLANs {
		vlans = append(vlans, VLANResourceModel{
			VLAN:      types.Int64Value(vlan.VLAN),
			Interface: types.StringValue(vlan.Interface),
			DHCP:      types.BoolValue(vlan.DHCP),
		})
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vlan"), vlans)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bridges []BridgeResourceModel
	for _, bridge := range result.NetworkProfile.Bridges {
		bridges = append(bridges, BridgeResourceModel{
			Name:      types.StringValue(bridge.Name),
			Interface: types.StringValue(bridge.Interface),
			DHCP:      types.BoolValue(bridge.DHCP),
		})
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bridge"), bridges)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bonds []BondResourceModel
	for _, bond := range result.NetworkProfile.Bonds {
		bonds = append(bonds, BondResourceModel{
			Name:       types.StringValue(bond.Name),
			Interfaces: types.StringValue(bond.Interfaces),
			Kind:       types.StringValue(bond.Kind),
			DHCP:       types.BoolValue(bond.DHCP),
		})
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bond"), bonds)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *NetworkProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var vlans []VLANResourceModel
	diags := data.VLANs.ElementsAs(ctx, &vlans, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	var bridges []BridgeResourceModel
	diags = data.Bridges.ElementsAs(ctx, &bridges, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	var bonds []BondResourceModel
	diags = data.Bonds.ElementsAs(ctx, &bonds, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	networkProfileUpdateInput := &client.NetworkProfileUpdateInput{
		Name:    data.Name.ValueString(),
		VLANs:   r.vlansInput(vlans),
		Bridges: r.bridgesInput(bridges),
		Bonds:   r.bondsInput(bonds),
	}

	result, err := r.client.NetworkProfile().Update(ctx, data.ID.ValueInt64(), networkProfileUpdateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update network profile, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update network profile, got error: %s", err))
			return
		}

		_, err := r.client.NetworkProfile().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network profile, got error: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

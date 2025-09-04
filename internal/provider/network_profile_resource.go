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
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Links types.List   `tfsdk:"link"`
}

type LinkResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Interfaces     types.List   `tfsdk:"interfaces"`
	VLANs          types.List   `tfsdk:"vlans"`
	HostNetworking types.Bool   `tfsdk:"host_networking"`
	LACP           types.Bool   `tfsdk:"lacp"`
	NativeVLAN     types.Bool   `tfsdk:"native_vlan"`
}

func (r *NetworkProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_profile"
}

const networkProfileDesc = "A [Network Profile](https://docs.meltcloud.io/concepts/networking/network-profiles) specifies the network configuration for [Machines](https://docs.meltcloud.io/concepts/machines) in a [Machine Pool](https://docs.meltcloud.io/guides/machine-pools/create.html)."

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

func linkResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Link name",
		},
		"interfaces": schema.ListAttribute{
			Required:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "List of interface names",
		},
		"vlans": schema.ListAttribute{
			Required:            true,
			ElementType:         types.Int64Type,
			MarkdownDescription: "List of VLAN IDs",
		},
		"host_networking": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether to use host networking",
		},
		"lacp": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether to use LACP (Link Aggregation Control Protocol)",
		},
		"native_vlan": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Whether to use the native VLAN",
		},
	}
}

func (r *NetworkProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: networkProfileDesc,

		Attributes: networkProfileResourceAttributes(),

		Blocks: map[string]schema.Block{
			"link": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: linkResourceAttributes(),
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

	var links []LinkResourceModel
	diags := data.Links.ElementsAs(ctx, &links, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	networkProfileCreateInput := &client.NetworkProfileCreateInput{
		Name:  data.Name.ValueString(),
		Links: r.linksInput(ctx, links),
	}

	result, err := r.client.NetworkProfile().Create(ctx, networkProfileCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create network profile, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.NetworkProfile.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkProfileResource) linksInput(ctx context.Context, links []LinkResourceModel) []client.Link {
	var linksInput []client.Link
	for _, linkConfiguration := range links {
		var interfaces []string
		linkConfiguration.Interfaces.ElementsAs(ctx, &interfaces, false)

		var vlans []int64
		linkConfiguration.VLANs.ElementsAs(ctx, &vlans, false)

		linksInput = append(linksInput, client.Link{
			Name:           linkConfiguration.Name.ValueString(),
			Interfaces:     interfaces,
			VLANs:          vlans,
			HostNetworking: linkConfiguration.HostNetworking.ValueBool(),
			LACP:           linkConfiguration.LACP.ValueBool(),
			NativeVLAN:     linkConfiguration.NativeVLAN.ValueBool(),
		})
	}
	return linksInput
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

	var links []LinkResourceModel
	for _, link := range result.NetworkProfile.Links {
		interfacesList, diags := types.ListValueFrom(ctx, types.StringType, link.Interfaces)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		vlansList, diags := types.ListValueFrom(ctx, types.Int64Type, link.VLANs)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		links = append(links, LinkResourceModel{
			Name:           types.StringValue(link.Name),
			Interfaces:     interfacesList,
			VLANs:          vlansList,
			HostNetworking: types.BoolValue(link.HostNetworking),
			LACP:           types.BoolValue(link.LACP),
			NativeVLAN:     types.BoolValue(link.NativeVLAN),
		})
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("link"), links)...)
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

	var links []LinkResourceModel
	diags := data.Links.ElementsAs(ctx, &links, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	networkProfileUpdateInput := &client.NetworkProfileUpdateInput{
		Name:  data.Name.ValueString(),
		Links: r.linksInput(ctx, links),
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

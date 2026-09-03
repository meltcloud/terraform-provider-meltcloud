package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

type SubnetResource struct {
	client *client.Client
}

type SubnetResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	NetworkID  types.Int64  `tfsdk:"network_id"`
	Name       types.String `tfsdk:"name"`
	VLAN       types.Int64  `tfsdk:"vlan"`
	Addressing types.String `tfsdk:"addressing"`
	IPPoolID   types.Int64  `tfsdk:"ip_pool_id"`
	Gateway    types.String `tfsdk:"gateway"`
	DNS        types.List   `tfsdk:"dns"`
	NTP        types.List   `tfsdk:"ntp"`
	Domains    types.List   `tfsdk:"domains"`
	MTU        types.Int64  `tfsdk:"mtu"`
	Routes     types.List   `tfsdk:"route"`
}

type RouteResourceModel struct {
	Destination types.String `tfsdk:"destination"`
	Via         types.String `tfsdk:"via"`
	Metric      types.Int64  `tfsdk:"metric"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

const subnetDesc = "A [Subnet](https://docs.meltcloud.io/concepts/networking) is one segment of a Network: which VLAN it is, how a Machine gets an address on it, and what it delivers besides the address."

func subnetResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the subnet on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"network_id": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "ID of the Network this subnet belongs to",
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the subnet, unique within its Network",
		},
		"vlan": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "VLAN ID of the segment. Leave empty when the segment carries no VLAN",
		},
		"addressing": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "How a Machine gets an address: `dhcp`, where an existing DHCP server provides them, or `ipam`, where meltcloud does",
		},
		"ip_pool_id": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "ID of the IP Pool the addresses come from. Required with addressing `ipam`",
		},
		"gateway": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The default route, configured only where the Host Network is primary. Only with addressing `ipam`: a DHCP server delivers its own",
		},
		"dns": schema.ListAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The resolvers to configure. With `dhcp`, setting these replaces what the server sends in option 6",
		},
		"ntp": schema.ListAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The time servers to configure. With `dhcp`, setting these replaces what the server sends in option 42",
		},
		"domains": schema.ListAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The search domains to configure. With `dhcp`, setting these replaces what the server sends in options 15 and 119",
		},
		"mtu": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The MTU to configure on the device. With `dhcp`, setting this replaces what the server sends in option 26",
		},
	}
}

func routeResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"destination": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The network this route reaches, in CIDR notation",
		},
		"via": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The router that reaches it",
		},
		"metric": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Lower wins when several routes match",
		},
	}
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: subnetDesc,
		Attributes:          subnetResourceAttributes(),
		Blocks: map[string]schema.Block{
			"route": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: routeResourceAttributes(),
				},
			},
		},
	}
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.SubnetCreateInput{
		Name:       data.Name.ValueString(),
		VLAN:       data.VLAN.ValueInt64Pointer(),
		Addressing: data.Addressing.ValueString(),
		IPPoolID:   data.IPPoolID.ValueInt64Pointer(),
		Gateway:    data.Gateway.ValueStringPointer(),
		DNS:        stringList(ctx, data.DNS),
		NTP:        stringList(ctx, data.NTP),
		Domains:    stringList(ctx, data.Domains),
		MTU:        data.MTU.ValueInt64Pointer(),
		Routes:     r.routesInput(ctx, data.Routes),
	}

	result, err := r.client.Subnet().Create(ctx, data.NetworkID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create subnet, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.Subnet.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func stringList(ctx context.Context, list types.List) []string {
	var values []string
	list.ElementsAs(ctx, &values, false)
	return values
}

func (r *SubnetResource) routesInput(ctx context.Context, list types.List) []client.SubnetRoute {
	var routes []RouteResourceModel
	list.ElementsAs(ctx, &routes, false)

	var input []client.SubnetRoute
	for _, route := range routes {
		input = append(input, client.SubnetRoute{
			Destination: route.Destination.ValueString(),
			Via:         route.Via.ValueString(),
			Metric:      route.Metric.ValueInt64Pointer(),
		})
	}
	return input
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Subnet().Get(ctx, data.NetworkID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read subnet, got error: %s", err))
		return
	}

	subnet := result.Subnet
	data.Name = types.StringValue(subnet.Name)
	data.VLAN = types.Int64PointerValue(subnet.VLAN)
	data.Addressing = types.StringValue(subnet.Addressing)
	data.IPPoolID = types.Int64PointerValue(subnet.IPPoolID)
	data.Gateway = types.StringPointerValue(subnet.Gateway)
	data.MTU = types.Int64PointerValue(subnet.MTU)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// A subnet is immutable in foundry: what it delivers reaches a machine, so a
// change is a new subnet.
func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Supported", "A subnet cannot be changed; it is replaced.")
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Subnet().Delete(ctx, data.NetworkID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete subnet, got error: %s", err))
		return
	}
}

var subnetImportIDPattern = regexp.MustCompile(`networks/(\d+)/subnets/(\d+)`)

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := subnetImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 3 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", subnetImportIDPattern.String()))
		return
	}

	networkID, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid Network ID: %s", err))
		return
	}

	id, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_id"), networkID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

package provider

import (
	"context"

	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the subnet, unique within its Network",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"vlan": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "VLAN ID of the segment. Leave empty when the segment carries no VLAN",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"addressing": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "How a Machine gets an address: `dhcp`, where an existing DHCP server provides them, or `ipam`, where meltcloud does",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"ip_pool_id": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "ID of the IP Pool the addresses come from. Required with addressing `ipam`",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"gateway": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The default route, configured only where the Host Network is primary. Only with addressing `ipam`: a DHCP server delivers its own",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"dns": schema.ListAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The resolvers to configure. With `dhcp`, setting these replaces what the server sends in option 6",
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		"ntp": schema.ListAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The time servers to configure. With `dhcp`, setting these replaces what the server sends in option 42",
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		"domains": schema.ListAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The search domains to configure. With `dhcp`, setting these replaces what the server sends in options 15 and 119",
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		"mtu": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The MTU to configure on the device. With `dhcp`, setting this replaces what the server sends in option 26",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
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
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
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
		VLAN:       int64Value(data.VLAN),
		Addressing: data.Addressing.ValueString(),
		IPPoolID:   int64Value(data.IPPoolID),
		Gateway:    stringValue(data.Gateway),
		DNS:        stringList(ctx, data.DNS),
		NTP:        stringList(ctx, data.NTP),
		Domains:    stringList(ctx, data.Domains),
		MTU:        int64Value(data.MTU),
		Routes:     r.routesInput(ctx, data.Routes),
	}

	result, err := r.client.Subnet().Create(ctx, data.NetworkID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create subnet, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(applySubnet(ctx, &data, result.Subnet)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// applySubnet takes what the server made of the subnet, which is what every
// attribute it fills in has to end up as.
func applySubnet(ctx context.Context, data *SubnetResourceModel, subnet *client.Subnet) diag.Diagnostics {
	var diags diag.Diagnostics

	dns, dnsDiags := stringListValue(ctx, subnet.DNS)
	diags.Append(dnsDiags...)
	ntp, ntpDiags := stringListValue(ctx, subnet.NTP)
	diags.Append(ntpDiags...)
	domains, domainDiags := stringListValue(ctx, subnet.Domains)
	diags.Append(domainDiags...)
	if diags.HasError() {
		return diags
	}

	data.ID = types.Int64Value(subnet.ID)
	data.NetworkID = types.Int64Value(subnet.NetworkID)
	data.Name = types.StringValue(subnet.Name)
	data.VLAN = types.Int64PointerValue(subnet.VLAN)
	data.Addressing = types.StringValue(subnet.Addressing)
	data.IPPoolID = types.Int64PointerValue(subnet.IPPoolID)
	data.Gateway = types.StringPointerValue(subnet.Gateway)
	data.MTU = types.Int64PointerValue(subnet.MTU)
	data.DNS = dns
	data.NTP = ntp
	data.Domains = domains

	return diags
}

// An empty list is how the server says a field was never set, and an attribute
// that was never set has to read as absent rather than as an empty list.
func stringListValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(types.StringType), nil
	}
	return types.ListValueFrom(ctx, types.StringType, values)
}

// An optional attribute the server fills in is unknown while planning, and an
// unknown value reads as zero rather than as absent.
func int64Value(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return value.ValueInt64Pointer()
}

func stringValue(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return value.ValueStringPointer()
}

func stringList(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

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
			Metric:      int64Value(route.Metric),
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

	resp.Diagnostics.Append(applySubnet(ctx, &data, result.Subnet)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

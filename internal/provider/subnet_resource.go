package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &SubnetResource{}
	_ resource.ResourceWithImportState    = &SubnetResource{}
	_ resource.ResourceWithValidateConfig = &SubnetResource{}
)

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
	Addressing types.String `tfsdk:"addressing"`
	VLAN       types.Int64  `tfsdk:"vlan"`
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

const subnetDesc = "A [Subnet](https://docs.meltcloud.io/concepts/networking/networks-and-subnets) is one segment of a Network. It defines how Machines get an address on it (DHCP or IPAM) and further network configuration (DNS servers, NTP servers, MTU, routes, ...)."

func subnetResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Subnet on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"network_id": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "ID of the Network this Subnet belongs to",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the Subnet, unique within its Network",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"addressing": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf("dhcp", "ipam"),
			},
			MarkdownDescription: "How a Machine gets an address: `dhcp`, where an existing DHCP server provides them, or `ipam`, where meltcloud does",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"vlan": schema.Int64Attribute{
			Optional: true,
			Validators: []validator.Int64{
				int64validator.Between(1, 4094),
			},
			MarkdownDescription: "VLAN ID of the segment. Leave empty when the segment has no VLAN",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"ip_pool_id": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "ID of the IP Pool addresses come from. Required with addressing `ipam`",
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
			Optional: true,
			Validators: []validator.Int64{
				int64validator.Between(1000, 9216),
			},
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

func (r *SubnetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() || data.Addressing.IsUnknown() {
		return
	}

	switch data.Addressing.ValueString() {
	case "ipam":
		if data.IPPoolID.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("ip_pool_id"),
				"Missing Attribute Configuration",
				"ip_pool_id must be set on an ipam subnet, which hands out addresses from that pool.",
			)
		}
	case "dhcp":
		if !data.IPPoolID.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("ip_pool_id"),
				"Invalid Attribute Combination",
				"ip_pool_id is only allowed on an ipam subnet.",
			)
		}

		if !data.Gateway.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("gateway"),
				"Invalid Attribute Combination",
				"gateway is only allowed on an ipam subnet, a dhcp subnet learns it from the wire.",
			)
		}
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
		Addressing: data.Addressing.ValueString(),
		VLAN:       int64Value(data.VLAN),
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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Subnet, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(applySubnet(ctx, &data, result.Subnet)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func applySubnet(ctx context.Context, data *SubnetResourceModel, subnet *client.Subnet) diag.Diagnostics {
	var diags diag.Diagnostics

	dns, dnsDiags := stringListValue(ctx, subnet.DNS)
	diags.Append(dnsDiags...)
	ntp, ntpDiags := stringListValue(ctx, subnet.NTP)
	diags.Append(ntpDiags...)
	domains, domainDiags := stringListValue(ctx, subnet.Domains)
	diags.Append(domainDiags...)
	routes, routeDiags := routeValues(ctx, subnet.Routes)
	diags.Append(routeDiags...)
	if diags.HasError() {
		return diags
	}

	data.ID = types.Int64Value(subnet.ID)
	data.NetworkID = types.Int64Value(subnet.NetworkID)
	data.Name = types.StringValue(subnet.Name)
	data.Addressing = types.StringValue(subnet.Addressing)
	data.VLAN = types.Int64PointerValue(subnet.VLAN)
	data.IPPoolID = types.Int64PointerValue(subnet.IPPoolID)
	data.Gateway = types.StringPointerValue(subnet.Gateway)
	data.MTU = types.Int64PointerValue(subnet.MTU)
	data.DNS = dns
	data.NTP = ntp
	data.Domains = domains
	data.Routes = routes

	return diags
}

func routeObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"destination": types.StringType,
		"via":         types.StringType,
		"metric":      types.Int64Type,
	}}
}

func routeValues(ctx context.Context, routes []client.SubnetRoute) (types.List, diag.Diagnostics) {
	if len(routes) == 0 {
		return types.ListNull(routeObjectType()), nil
	}

	models := make([]RouteResourceModel, 0, len(routes))
	for _, route := range routes {
		models = append(models, RouteResourceModel{
			Destination: types.StringValue(route.Destination),
			Via:         types.StringValue(route.Via),
			Metric:      types.Int64PointerValue(route.Metric),
		})
	}

	return types.ListValueFrom(ctx, routeObjectType(), models)
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

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Subnet, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(applySubnet(ctx, &data, result.Subnet)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Supported", "A Subnet cannot be changed; it is replaced.")
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Subnet().Delete(ctx, data.NetworkID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Subnet, got error: %s", err))
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

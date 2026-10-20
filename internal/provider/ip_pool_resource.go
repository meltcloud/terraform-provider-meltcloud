package provider

import (
	"context"
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                   = &IPPoolResource{}
	_ resource.ResourceWithImportState    = &IPPoolResource{}
	_ resource.ResourceWithValidateConfig = &IPPoolResource{}
)

func NewIPPoolResource() resource.Resource {
	return &IPPoolResource{}
}

type IPPoolResource struct {
	client *client.Client
}

type IPPoolResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CIDR        types.String `tfsdk:"cidr"`
	Description types.String `tfsdk:"description"`
	Ranges      types.Set    `tfsdk:"ranges"`
}

type IPPoolRangeResourceModel struct {
	Kind         types.String `tfsdk:"kind"`
	StartAddress types.String `tfsdk:"start_address"`
	EndAddress   types.String `tfsdk:"end_address"`
	Description  types.String `tfsdk:"description"`
}

func (r *IPPoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_pool"
}

const ipPoolDesc = "An [IP Pool](https://docs.meltcloud.io/concepts/networking/subnet-addressing#ipam) hands out addresses from a CIDR. It is used by [Subnets](https://docs.meltcloud.io/concepts/networking/networks-and-subnets) with addressing `ipam`."

func ipPoolResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the IP Pool on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the IP Pool, unique within the organization",
		},
		"cidr": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The CIDR addresses come from. It has to be the prefix of the segment the IP Pool serves, and cannot be changed",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "What this IP Pool is for",
		},
		"ranges": schema.SetNestedAttribute{
			Optional:            true,
			MarkdownDescription: "The ranges of the CIDR that addresses are allocated from, or kept free",
			NestedObject: schema.NestedAttributeObject{
				Attributes: ipPoolRangeResourceAttributes(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
	}
}

func ipPoolRangeResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"kind": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf("allocatable", "excluded"),
			},
			MarkdownDescription: "`allocatable`, which addresses are taken from, or `excluded`, which keeps the addresses inside one of them free",
		},
		"start_address": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The first address the range covers",
		},
		"end_address": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The last address the range covers",
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "What sits here",
		},
	}
}

func (r *IPPoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ipPoolDesc,
		Attributes:          ipPoolResourceAttributes(),
	}
}

func (r *IPPoolResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data IPPoolResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() || data.Ranges.IsNull() || data.Ranges.IsUnknown() {
		return
	}

	for _, element := range data.Ranges.Elements() {
		var poolRange IPPoolRangeResourceModel
		object, ok := element.(types.Object)
		if !ok {
			continue
		}
		resp.Diagnostics.Append(object.As(ctx, &poolRange, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		validateRange(resp, path.Root("ranges").AtSetValue(element), poolRange)
	}
}

func validateRange(resp *resource.ValidateConfigResponse, rangePath path.Path, poolRange IPPoolRangeResourceModel) {
	start, startOK := parseAddress(resp, rangePath.AtName("start_address"), poolRange.StartAddress)
	end, endOK := parseAddress(resp, rangePath.AtName("end_address"), poolRange.EndAddress)

	if startOK && endOK && end.Less(start) {
		resp.Diagnostics.AddAttributeError(
			rangePath.AtName("end_address"),
			"Invalid Attribute Value",
			"end_address must not be before start_address.",
		)
	}
}

func parseAddress(resp *resource.ValidateConfigResponse, addressPath path.Path, value types.String) (netip.Addr, bool) {
	if value.IsNull() || value.IsUnknown() {
		return netip.Addr{}, false
	}

	address, err := netip.ParseAddr(value.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(addressPath, "Invalid Attribute Value", "Must be an IP address.")
		return netip.Addr{}, false
	}

	return address, true
}

func (r *IPPoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func ipPoolRangeObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"kind":          types.StringType,
		"start_address": types.StringType,
		"end_address":   types.StringType,
		"description":   types.StringType,
	}}
}

func ipPoolRangeValues(ctx context.Context, ranges []client.IPPoolRange) (types.Set, diag.Diagnostics) {
	if len(ranges) == 0 {
		return types.SetNull(ipPoolRangeObjectType()), nil
	}

	models := make([]IPPoolRangeResourceModel, 0, len(ranges))
	for _, poolRange := range ranges {
		models = append(models, IPPoolRangeResourceModel{
			Kind:         types.StringValue(poolRange.Kind),
			StartAddress: types.StringValue(poolRange.StartAddress),
			EndAddress:   types.StringValue(poolRange.EndAddress),
			Description:  types.StringPointerValue(poolRange.Description),
		})
	}

	return types.SetValueFrom(ctx, ipPoolRangeObjectType(), models)
}

func (r *IPPoolResource) rangesInput(ctx context.Context, set types.Set) []client.IPPoolRange {
	var ranges []IPPoolRangeResourceModel
	set.ElementsAs(ctx, &ranges, false)

	// Empty rather than nil: the update sends every field, and a null list leaves
	// the ranges the IP Pool has.
	input := []client.IPPoolRange{}
	for _, poolRange := range ranges {
		input = append(input, client.IPPoolRange{
			Kind:         poolRange.Kind.ValueString(),
			StartAddress: poolRange.StartAddress.ValueString(),
			EndAddress:   poolRange.EndAddress.ValueString(),
			Description:  stringValue(poolRange.Description),
		})
	}
	return input
}

func (r *IPPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.IPPoolCreateInput{
		Name:        data.Name.ValueString(),
		CIDR:        data.CIDR.ValueString(),
		Description: stringValue(data.Description),
		Ranges:      r.rangesInput(ctx, data.Ranges),
	}

	result, err := r.client.IPPool().Create(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IP Pool, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.IPPool.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.IPPool().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IP Pool, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.IPPool.Name)
	data.CIDR = types.StringValue(result.IPPool.CIDR)
	data.Description = types.StringPointerValue(result.IPPool.Description)

	ranges, diags := ipPoolRangeValues(ctx, result.IPPool.Ranges)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Ranges = ranges

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IPPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.IPPoolUpdateInput{
		Name:        data.Name.ValueString(),
		Description: stringValue(data.Description),
		Ranges:      r.rangesInput(ctx, data.Ranges),
	}

	_, err := r.client.IPPool().Update(ctx, data.ID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IP Pool, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.IPPool().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IP Pool, got error: %s", err))
		return
	}
}

var ipPoolImportIDPattern = regexp.MustCompile(`ip_pools/(\d+)`)

func (r *IPPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := ipPoolImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", ipPoolImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

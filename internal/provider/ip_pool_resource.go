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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IPPoolResource{}
var _ resource.ResourceWithImportState = &IPPoolResource{}

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
	Ranges      types.List   `tfsdk:"range"`
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

const ipPoolDesc = "An [IP Pool](https://docs.meltcloud.io/concepts/networking) hands out addresses from a CIDR. A Subnet with addressing `ipam` draws from one."

func ipPoolResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the IP pool on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the IP pool, unique within the organization",
		},
		"cidr": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The network addresses are handed out from. Every address carries its prefix, so it cannot be changed: it has to be the network of the segment the pool serves",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "What this IP pool is for",
		},
	}
}

func ipPoolRangeResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"kind": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "`allocatable`, where addresses are handed out from, or `excluded`, a hole inside one",
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
		Blocks: map[string]schema.Block{
			"range": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: ipPoolRangeResourceAttributes(),
				},
			},
		},
	}
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

func (r *IPPoolResource) rangesInput(ctx context.Context, list types.List) []client.IPPoolRange {
	var ranges []IPPoolRangeResourceModel
	list.ElementsAs(ctx, &ranges, false)

	var input []client.IPPoolRange
	for _, poolRange := range ranges {
		input = append(input, client.IPPoolRange{
			Kind:         poolRange.Kind.ValueString(),
			StartAddress: poolRange.StartAddress.ValueString(),
			EndAddress:   poolRange.EndAddress.ValueString(),
			Description:  poolRange.Description.ValueStringPointer(),
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
		Description: data.Description.ValueStringPointer(),
		Ranges:      r.rangesInput(ctx, data.Ranges),
	}

	result, err := r.client.IPPool().Create(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IP pool, got error: %s", err))
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

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IP pool, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.IPPool.Name)
	data.CIDR = types.StringValue(result.IPPool.CIDR)
	data.Description = types.StringPointerValue(result.IPPool.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// The ranges sent are the ranges the pool ends up with, so a plan describes the
// pool it wants rather than the changes to it.
func (r *IPPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IPPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.IPPoolUpdateInput{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueStringPointer(),
		Ranges:      r.rangesInput(ctx, data.Ranges),
	}

	_, err := r.client.IPPool().Update(ctx, data.ID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IP pool, got error: %s", err))
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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IP pool, got error: %s", err))
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

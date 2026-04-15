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

var _ resource.Resource = &ElasticShareResource{}
var _ resource.ResourceWithImportState = &ElasticShareResource{}

func NewElasticShareResource() resource.Resource {
	return &ElasticShareResource{}
}

type ElasticShareResource struct {
	client *client.Client
}

type ElasticShareResourceModel struct {
	ID                      types.Int64  `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Cores                   types.Int64  `tfsdk:"cores"`
	DiskGB                  types.Int64  `tfsdk:"disk_gb"`
	MemoryMB                types.Int64  `tfsdk:"memory_mb"`
	CapacityID              types.Int64  `tfsdk:"capacity_id"`
	ConsumingOrganizationUUID types.String `tfsdk:"consuming_organization_uuid"`
}

func (r *ElasticShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_share"
}

const elasticShareDesc = "An Elastic Share grants a consuming organization a slice of an Elastic Capacity, expressed in cores, memory, and disk."

func elasticShareResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Elastic Share on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Elastic Share",
			Required:            true,
		},
		"cores": schema.Int64Attribute{
			MarkdownDescription: "Number of cores granted to the consuming organization",
			Required:            true,
		},
		"disk_gb": schema.Int64Attribute{
			MarkdownDescription: "Disk in GB granted to the consuming organization",
			Required:            true,
		},
		"memory_mb": schema.Int64Attribute{
			MarkdownDescription: "Memory in MB granted to the consuming organization",
			Required:            true,
		},
		"capacity_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the parent Elastic Capacity",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"consuming_organization_uuid": schema.StringAttribute{
			MarkdownDescription: "UUID of the consuming Organization",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r *ElasticShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticShareDesc,
		Attributes:          elasticShareResourceAttributes(),
	}
}

func (r *ElasticShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *ElasticShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.ElasticShareCreateInput{
		Name:                    data.Name.ValueString(),
		Cores:                   data.Cores.ValueInt64(),
		DiskGB:                  data.DiskGB.ValueInt64(),
		MemoryMB:                data.MemoryMB.ValueInt64(),
		CapacityID:              data.CapacityID.ValueInt64(),
		ConsumingOrganizationUUID: data.ConsumingOrganizationUUID.ValueString(),
	}

	result, err := r.client.ElasticShare().Create(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic share, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.ElasticShare.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticShare().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic share, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.ElasticShare.Name)
	data.Cores = types.Int64Value(result.ElasticShare.Cores)
	data.DiskGB = types.Int64Value(result.ElasticShare.DiskGB)
	data.MemoryMB = types.Int64Value(result.ElasticShare.MemoryMB)
	data.CapacityID = types.Int64Value(result.ElasticShare.CapacityID)
	data.ConsumingOrganizationUUID = types.StringValue(result.ElasticShare.ConsumingOrganizationUUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ElasticShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.ElasticShareUpdateInput{
		Name:     data.Name.ValueString(),
		Cores:    data.Cores.ValueInt64(),
		DiskGB:   data.DiskGB.ValueInt64(),
		MemoryMB: data.MemoryMB.ValueInt64(),
	}

	_, err := r.client.ElasticShare().Update(ctx, data.ID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update elastic share, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ElasticShare().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic share, got error: %s", err))
		return
	}
}

var elasticShareImportIDPattern = regexp.MustCompile(`elastic_shares/(\d+)`)

func (r *ElasticShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := elasticShareImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", elasticShareImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

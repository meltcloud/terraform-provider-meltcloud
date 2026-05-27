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

var _ resource.Resource = &ElasticQuotaResource{}
var _ resource.ResourceWithImportState = &ElasticQuotaResource{}

func NewElasticQuotaResource() resource.Resource {
	return &ElasticQuotaResource{}
}

type ElasticQuotaResource struct {
	client *client.Client
}

type ElasticQuotaResourceModel struct {
	ID                        types.Int64  `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Cores                     types.Int64  `tfsdk:"cores"`
	DiskGB                    types.Int64  `tfsdk:"disk_gb"`
	MemoryMB                  types.Int64  `tfsdk:"memory_mb"`
	ElasticFleetID            types.Int64  `tfsdk:"elastic_fleet_id"`
	ConsumingOrganizationUUID types.String `tfsdk:"consuming_organization_uuid"`
}

func (r *ElasticQuotaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_quota"
}

const elasticQuotaDesc = "An [Elastic Quota](https://docs.meltcloud.io/tasks/elastic-fleets/create-quota) allocates a portion of an [Elastic Fleet](https://docs.meltcloud.io/concepts/elastic-node-pools#elastic-fleet)'s resources (CPU, RAM, disk) to a consuming organization."

func elasticQuotaResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Elastic Quota on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Elastic Quota",
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
		"elastic_fleet_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the parent Elastic Fleet",
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

func (r *ElasticQuotaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticQuotaDesc,
		Attributes:          elasticQuotaResourceAttributes(),
	}
}

func (r *ElasticQuotaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticQuotaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticQuotaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.ElasticQuotaCreateInput{
		Name:                      data.Name.ValueString(),
		Cores:                     data.Cores.ValueInt64(),
		DiskGB:                    data.DiskGB.ValueInt64(),
		MemoryMB:                  data.MemoryMB.ValueInt64(),
		ElasticFleetID:            data.ElasticFleetID.ValueInt64(),
		ConsumingOrganizationUUID: data.ConsumingOrganizationUUID.ValueString(),
	}

	result, err := r.client.ElasticQuota().Create(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic quota, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.ElasticQuota.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticQuotaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticQuotaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticQuota().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic quota, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.ElasticQuota.Name)
	data.Cores = types.Int64Value(result.ElasticQuota.Cores)
	data.DiskGB = types.Int64Value(result.ElasticQuota.DiskGB)
	data.MemoryMB = types.Int64Value(result.ElasticQuota.MemoryMB)
	data.ElasticFleetID = types.Int64Value(result.ElasticQuota.ElasticFleetID)
	data.ConsumingOrganizationUUID = types.StringValue(result.ElasticQuota.ConsumingOrganizationUUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticQuotaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ElasticQuotaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.ElasticQuotaUpdateInput{
		Name:     data.Name.ValueString(),
		Cores:    data.Cores.ValueInt64(),
		DiskGB:   data.DiskGB.ValueInt64(),
		MemoryMB: data.MemoryMB.ValueInt64(),
	}

	_, err := r.client.ElasticQuota().Update(ctx, data.ID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update elastic quota, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticQuotaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticQuotaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ElasticQuota().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic quota, got error: %s", err))
		return
	}
}

var elasticQuotaImportIDPattern = regexp.MustCompile(`elastic_quotas/(\d+)`)

func (r *ElasticQuotaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := elasticQuotaImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", elasticQuotaImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

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

var _ resource.Resource = &ElasticCapacityResource{}
var _ resource.ResourceWithImportState = &ElasticCapacityResource{}

func NewElasticCapacityResource() resource.Resource {
	return &ElasticCapacityResource{}
}

type ElasticCapacityResource struct {
	client *client.Client
}

type ElasticCapacityResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ClusterID types.Int64  `tfsdk:"cluster_id"`
	Status    types.String `tfsdk:"status"`
}

func (r *ElasticCapacityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_capacity"
}

const elasticCapacityDesc = "An Elastic Capacity is a pool of compute resources on a cluster which can be sliced into Elastic Shares for consumption by other organizations."

func elasticCapacityResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Elastic Capacity on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Elastic Capacity",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"cluster_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the associated cluster",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"status": schema.StringAttribute{
			MarkdownDescription: "Status of the Elastic Capacity",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func (r *ElasticCapacityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticCapacityDesc,
		Attributes:          elasticCapacityResourceAttributes(),
	}
}

func (r *ElasticCapacityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticCapacityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticCapacityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.ElasticCapacityCreateInput{
		Name:      data.Name.ValueString(),
		ClusterID: data.ClusterID.ValueInt64(),
	}

	result, err := r.client.ElasticCapacity().Create(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic capacity, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.ElasticCapacity.ID)

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic capacity, got error: %s", err))
			return
		}

		getResult, err := r.client.ElasticCapacity().Get(ctx, result.ElasticCapacity.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic capacity, got error: %s", err))
			return
		}
		data.Status = types.StringValue(getResult.ElasticCapacity.Status)
	} else {
		data.Status = types.StringValue(result.ElasticCapacity.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticCapacityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticCapacityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticCapacity().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic capacity, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.ElasticCapacity.Name)
	data.ClusterID = types.Int64Value(result.ElasticCapacity.ClusterID)
	data.Status = types.StringValue(result.ElasticCapacity.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticCapacityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are RequiresReplace; no in-place update.
}

func (r *ElasticCapacityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticCapacityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticCapacity().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic capacity, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic capacity, got error: %s", err))
			return
		}
	}
}

var elasticCapacityImportIDPattern = regexp.MustCompile(`elastic_capacities/(\d+)`)

func (r *ElasticCapacityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := elasticCapacityImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", elasticCapacityImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

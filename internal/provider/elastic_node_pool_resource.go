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

var _ resource.Resource = &ElasticNodePoolResource{}
var _ resource.ResourceWithImportState = &ElasticNodePoolResource{}

func NewElasticNodePoolResource() resource.Resource {
	return &ElasticNodePoolResource{}
}

type ElasticNodePoolResource struct {
	client *client.Client
}

type ElasticNodePoolResourceModel struct {
	ID             types.Int64      `tfsdk:"id"`
	ClusterID      types.Int64      `tfsdk:"cluster_id"`
	Name           types.String     `tfsdk:"name"`
	ElasticQuotaID types.Int64      `tfsdk:"elastic_quota_id"`
	Version        types.String     `tfsdk:"version"`
	PatchVersion   types.String     `tfsdk:"patch_version"`
	NodeCount      types.Int64      `tfsdk:"node_count"`
	Status         types.String     `tfsdk:"status"`
	NodeConfig     *NodeConfigModel `tfsdk:"node_config"`
}

type NodeConfigModel struct {
	Cores    types.Int64 `tfsdk:"cores"`
	MemoryMB types.Int64 `tfsdk:"memory_mb"`
	DiskGB   types.Int64 `tfsdk:"disk_gb"`
}

func (r *ElasticNodePoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_node_pool"
}

const elasticNodePoolDesc = "An Elastic Node Pool is a grouping of Elastic Fleet nodes scheduled into a consuming organization's cluster, sized via node count and per-node resources."

func elasticNodePoolResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Elastic Node Pool on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"cluster_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the cluster the node pool runs on",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Elastic Node Pool",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"elastic_quota_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Elastic Quota backing the node pool",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes minor version of the Elastic Node Pool nodes (Kubelet)",
			Required:            true,
		},
		"patch_version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes patch version of the Elastic Node Pool nodes (Kubelet)",
			Computed:            true,
		},
		"node_count": schema.Int64Attribute{
			MarkdownDescription: "Number of nodes in the node pool",
			Required:            true,
		},
		"status": schema.StringAttribute{
			MarkdownDescription: "Status of the Elastic Node Pool",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func nodeConfigBlockAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"cores": schema.Int64Attribute{
			MarkdownDescription: "Number of cores per node",
			Required:            true,
		},
		"memory_mb": schema.Int64Attribute{
			MarkdownDescription: "Memory in MB per node",
			Required:            true,
		},
		"disk_gb": schema.Int64Attribute{
			MarkdownDescription: "Disk in GB per node",
			Required:            true,
		},
	}
}

func (r *ElasticNodePoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticNodePoolDesc,
		Attributes:          elasticNodePoolResourceAttributes(),
		Blocks: map[string]schema.Block{
			"node_config": schema.SingleNestedBlock{
				MarkdownDescription: "Per-node resource configuration",
				Attributes:          nodeConfigBlockAttributes(),
			},
		},
	}
}

func (r *ElasticNodePoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticNodePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticNodePoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.NodeConfig == nil {
		resp.Diagnostics.AddError("Config Error", "node_config block is required")
		return
	}

	input := &client.ElasticNodePoolCreateInput{
		Name:           data.Name.ValueString(),
		ElasticQuotaID: data.ElasticQuotaID.ValueInt64(),
		NodeCount:      data.NodeCount.ValueInt64(),
		NodeCores:      data.NodeConfig.Cores.ValueInt64(),
		NodeMemoryMB:   data.NodeConfig.MemoryMB.ValueInt64(),
		NodeDiskGB:     data.NodeConfig.DiskGB.ValueInt64(),
		Version:        data.Version.ValueString(),
	}

	result, err := r.client.ElasticNodePool().Create(ctx, data.ClusterID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic node pool, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.ElasticNodePool.ID)

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic node pool, got error: %s", err))
			return
		}

		getResult, err := r.client.ElasticNodePool().Get(ctx, data.ClusterID.ValueInt64(), result.ElasticNodePool.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic node pool, got error: %s", err))
			return
		}
		data.PatchVersion = types.StringValue(getResult.ElasticNodePool.PatchVersion)
		data.Status = types.StringValue(getResult.ElasticNodePool.Status)
	} else {
		data.PatchVersion = types.StringValue(result.ElasticNodePool.PatchVersion)
		data.Status = types.StringValue(result.ElasticNodePool.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticNodePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticNodePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticNodePool().Get(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic node pool, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.ElasticNodePool.Name)
	data.ElasticQuotaID = types.Int64Value(result.ElasticNodePool.ElasticQuotaID)
	data.NodeCount = types.Int64Value(result.ElasticNodePool.NodeCount)
	data.Version = types.StringValue(result.ElasticNodePool.Version)
	data.PatchVersion = types.StringValue(result.ElasticNodePool.PatchVersion)
	data.Status = types.StringValue(result.ElasticNodePool.Status)
	data.NodeConfig = &NodeConfigModel{
		Cores:    types.Int64Value(result.ElasticNodePool.NodeCores),
		MemoryMB: types.Int64Value(result.ElasticNodePool.NodeMemoryMB),
		DiskGB:   types.Int64Value(result.ElasticNodePool.NodeDiskGB),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticNodePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ElasticNodePoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.NodeConfig == nil {
		resp.Diagnostics.AddError("Config Error", "node_config block is required")
		return
	}

	input := &client.ElasticNodePoolUpdateInput{
		NodeCount:    data.NodeCount.ValueInt64(),
		NodeCores:    data.NodeConfig.Cores.ValueInt64(),
		NodeMemoryMB: data.NodeConfig.MemoryMB.ValueInt64(),
		NodeDiskGB:   data.NodeConfig.DiskGB.ValueInt64(),
		Version:      data.Version.ValueString(),
	}

	result, err := r.client.ElasticNodePool().Update(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update elastic node pool, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update elastic node pool, got error: %s", err))
			return
		}

		getResult, err := r.client.ElasticNodePool().Get(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic node pool, got error: %s", err))
			return
		}
		data.PatchVersion = types.StringValue(getResult.ElasticNodePool.PatchVersion)
		data.Status = types.StringValue(getResult.ElasticNodePool.Status)
	} else {
		data.PatchVersion = types.StringValue(result.ElasticNodePool.PatchVersion)
		data.Status = types.StringValue(result.ElasticNodePool.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticNodePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticNodePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticNodePool().Delete(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic node pool, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic node pool, got error: %s", err))
			return
		}
	}
}

var elasticNodePoolImportIDPattern = regexp.MustCompile(`clusters/(\d+)/elastic_node_pools/(\d+)`)

func (r *ElasticNodePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := elasticNodePoolImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 3 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", elasticNodePoolImportIDPattern.String()))
		return
	}

	clusterID, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid cluster ID: %s", err))
		return
	}

	id, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

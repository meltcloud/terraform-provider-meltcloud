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

var _ resource.Resource = &ElasticPoolResource{}
var _ resource.ResourceWithImportState = &ElasticPoolResource{}

func NewElasticPoolResource() resource.Resource {
	return &ElasticPoolResource{}
}

type ElasticPoolResource struct {
	client *client.Client
}

type ElasticPoolResourceModel struct {
	ID           types.Int64           `tfsdk:"id"`
	ClusterID    types.Int64           `tfsdk:"cluster_id"`
	Name         types.String          `tfsdk:"name"`
	ShareID      types.Int64           `tfsdk:"share_id"`
	Version      types.String          `tfsdk:"version"`
	PatchVersion types.String          `tfsdk:"patch_version"`
	NodeCount    types.Int64           `tfsdk:"node_count"`
	Status       types.String          `tfsdk:"status"`
	NodeConfig   *NodeConfigModel      `tfsdk:"node_config"`
}

type NodeConfigModel struct {
	Cores    types.Int64 `tfsdk:"cores"`
	MemoryMB types.Int64 `tfsdk:"memory_mb"`
	DiskGB   types.Int64 `tfsdk:"disk_gb"`
}

func (r *ElasticPoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_pool"
}

const elasticPoolDesc = "An Elastic Pool is a grouping of Elastic Capacity nodes scheduled into a consuming organization's cluster, sized via node count and per-node resources."

func elasticPoolResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Elastic Pool on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"cluster_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the cluster the pool runs on",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Elastic Pool",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"share_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Elastic Share backing the pool",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes minor version of the Elastic Pool nodes (Kubelet)",
			Required:            true,
		},
		"patch_version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes patch version of the Elastic Pool nodes (Kubelet)",
			Computed:            true,
		},
		"node_count": schema.Int64Attribute{
			MarkdownDescription: "Number of nodes in the pool",
			Required:            true,
		},
		"status": schema.StringAttribute{
			MarkdownDescription: "Status of the Elastic Pool",
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

func (r *ElasticPoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: elasticPoolDesc,
		Attributes:          elasticPoolResourceAttributes(),
		Blocks: map[string]schema.Block{
			"node_config": schema.SingleNestedBlock{
				MarkdownDescription: "Per-node resource configuration",
				Attributes:          nodeConfigBlockAttributes(),
			},
		},
	}
}

func (r *ElasticPoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.NodeConfig == nil {
		resp.Diagnostics.AddError("Config Error", "node_config block is required")
		return
	}

	input := &client.ElasticPoolCreateInput{
		Name:         data.Name.ValueString(),
		ShareID:      data.ShareID.ValueInt64(),
		NodeCount:    data.NodeCount.ValueInt64(),
		NodeCores:    data.NodeConfig.Cores.ValueInt64(),
		NodeMemoryMB: data.NodeConfig.MemoryMB.ValueInt64(),
		NodeDiskGB:   data.NodeConfig.DiskGB.ValueInt64(),
		Version:      data.Version.ValueString(),
	}

	result, err := r.client.ElasticPool().Create(ctx, data.ClusterID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic pool, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.ElasticPool.ID)

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create elastic pool, got error: %s", err))
			return
		}

		getResult, err := r.client.ElasticPool().Get(ctx, data.ClusterID.ValueInt64(), result.ElasticPool.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic pool, got error: %s", err))
			return
		}
		data.PatchVersion = types.StringValue(getResult.ElasticPool.PatchVersion)
		data.Status = types.StringValue(getResult.ElasticPool.Status)
	} else {
		data.PatchVersion = types.StringValue(result.ElasticPool.PatchVersion)
		data.Status = types.StringValue(result.ElasticPool.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticPool().Get(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic pool, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.ElasticPool.Name)
	data.ShareID = types.Int64Value(result.ElasticPool.ShareID)
	data.NodeCount = types.Int64Value(result.ElasticPool.NodeCount)
	data.Version = types.StringValue(result.ElasticPool.Version)
	data.PatchVersion = types.StringValue(result.ElasticPool.PatchVersion)
	data.Status = types.StringValue(result.ElasticPool.Status)
	data.NodeConfig = &NodeConfigModel{
		Cores:    types.Int64Value(result.ElasticPool.NodeCores),
		MemoryMB: types.Int64Value(result.ElasticPool.NodeMemoryMB),
		DiskGB:   types.Int64Value(result.ElasticPool.NodeDiskGB),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ElasticPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.NodeConfig == nil {
		resp.Diagnostics.AddError("Config Error", "node_config block is required")
		return
	}

	input := &client.ElasticPoolUpdateInput{
		NodeCount:    data.NodeCount.ValueInt64(),
		NodeCores:    data.NodeConfig.Cores.ValueInt64(),
		NodeMemoryMB: data.NodeConfig.MemoryMB.ValueInt64(),
		NodeDiskGB:   data.NodeConfig.DiskGB.ValueInt64(),
		Version:      data.Version.ValueString(),
	}

	result, err := r.client.ElasticPool().Update(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64(), input)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update elastic pool, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update elastic pool, got error: %s", err))
			return
		}

		getResult, err := r.client.ElasticPool().Get(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read elastic pool, got error: %s", err))
			return
		}
		data.PatchVersion = types.StringValue(getResult.ElasticPool.PatchVersion)
		data.Status = types.StringValue(getResult.ElasticPool.Status)
	} else {
		data.PatchVersion = types.StringValue(result.ElasticPool.PatchVersion)
		data.Status = types.StringValue(result.ElasticPool.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ElasticPool().Delete(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic pool, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete elastic pool, got error: %s", err))
			return
		}
	}
}

var elasticPoolImportIDPattern = regexp.MustCompile(`clusters/(\d+)/elastic_pools/(\d+)`)

func (r *ElasticPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := elasticPoolImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 3 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", elasticPoolImportIDPattern.String()))
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

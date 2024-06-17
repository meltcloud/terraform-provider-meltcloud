package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"terraform-provider-melt/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MachinePoolResource{}
var _ resource.ResourceWithImportState = &MachinePoolResource{}

func NewMachinePoolResource() resource.Resource {
	return &MachinePoolResource{}
}

// MachinePoolResource defines the resource implementation.
type MachinePoolResource struct {
	client *client.Client
}

// MachinePoolResourceModel describes the resource data model.
type MachinePoolResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	ClusterId         types.Int64  `tfsdk:"cluster_id"`
	Name              types.String `tfsdk:"name"`
	PrimaryDiskDevice types.String `tfsdk:"primary_disk_device"`
	Version           types.String `tfsdk:"version"`
	PatchVersion      types.String `tfsdk:"patch_version"`
}

func (r *MachinePoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_pool"
}

func (r *MachinePoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MachinePool",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Machine Pool Melt ID",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.Int64Attribute{
				MarkdownDescription: "ID of the associated cluster",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the machine pool",
				Required:            true,
			},
			"primary_disk_device": schema.StringAttribute{
				MarkdownDescription: "Name of the primary disk of the machine, i.e. /dev/vda",
				Required:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes minor version of the machine pool",
				Required:            true,
			},
			"patch_version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes patch version of the machine pool",
				Computed:            true,
			},
		},
	}
}

func (r *MachinePoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *MachinePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MachinePoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	machinePoolCreateInput := &client.MachinePoolCreateInput{
		Name:              data.Name.ValueString(),
		PrimaryDiskDevice: data.PrimaryDiskDevice.ValueString(),
		UserVersion:       data.Version.ValueString(),
	}

	result, err := r.client.MachinePool().Create(ctx, data.ClusterId.ValueInt64(), machinePoolCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create machine pool, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.MachinePool.ID)
	data.PatchVersion = types.StringValue(result.MachinePool.PatchVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachinePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MachinePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.MachinePool().Get(ctx, data.ClusterId.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine pool, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.MachinePool.Name)
	data.Version = types.StringValue(result.MachinePool.UserVersion)
	data.PatchVersion = types.StringValue(result.MachinePool.PatchVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachinePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MachinePoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	machinePoolUpdateInput := &client.MachinePoolUpdateInput{
		Name:              data.Name.ValueString(),
		PrimaryDiskDevice: data.PrimaryDiskDevice.ValueString(),
		UserVersion:       data.Version.ValueString(),
	}

	result, err := r.client.MachinePool().Update(ctx, data.ClusterId.ValueInt64(), data.ID.ValueInt64(), machinePoolUpdateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine pool, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine pool, got error: %s", err))
			return
		}

		// TODO handle failed state

		_, err := r.client.MachinePool().Get(ctx, data.ClusterId.ValueInt64(), data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine pool, got error: %s", err))
			return
		}
		data.PatchVersion = types.StringValue(result.MachinePool.PatchVersion)
	} else {
		data.PatchVersion = types.StringValue(result.MachinePool.PatchVersion)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachinePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MachinePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.MachinePool().Delete(ctx, data.ClusterId.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete machine pool, got error: %s", err))
		return
	}
}

func (r *MachinePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

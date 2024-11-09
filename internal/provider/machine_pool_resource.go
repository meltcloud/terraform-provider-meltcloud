package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

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
	ID                         types.Int64  `tfsdk:"id"`
	ClusterId                  types.Int64  `tfsdk:"cluster_id"`
	Name                       types.String `tfsdk:"name"`
	PrimaryDiskDevice          types.String `tfsdk:"primary_disk_device"`
	ReuseExistingRootPartition types.Bool   `tfsdk:"reuse_existing_root_partition"`
	Version                    types.String `tfsdk:"version"`
	PatchVersion               types.String `tfsdk:"patch_version"`
	NetworkConfigurations      types.List   `tfsdk:"network_configuration"`
}

type NetworkConfigurationResourceModel struct {
	Type       types.String `tfsdk:"type"`
	Interfaces types.String `tfsdk:"interfaces"`
	VLANMode   types.String `tfsdk:"vlan_mode"`
	VLANs      types.String `tfsdk:"vlans"`
}

func (r *MachinePoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_pool"
}

const machinePoolDesc = "A [Machine Pool](https://meltcloud.io/docs/guides/machine-pools/create.html) is a grouping entity for Machines (Kubernetes workers) " +
	"which share a set of common configuration such as Kubelet version, disk or network configuration."

func machinePoolResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Machine Pool on meltcloud",
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
			Computed:            true,
			Optional:            true,
			Default:             stringdefault.StaticString(""),
		},
		"reuse_existing_root_partition": schema.BoolAttribute{
			MarkdownDescription: "Reuse existing Partition for the ephemeral root",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			Validators: []validator.Bool{
				boolvalidator.Equals(true),
				boolvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRoot("primary_disk_device"),
				}...),
			},
		},
		"version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes minor version of the machine pool (Kubelet)",
			Required:            true,
		},
		"patch_version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes patch version of the machine pool (Kubelet)",
			Computed:            true,
		},
	}
}

func networkConfigurationResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The network type - must be 'native' or 'bond'",
		},

		"interfaces": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Interface name (for network type native), wildcard or space-separated list of interfaces (for network type bond)",
		},

		"vlan_mode": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The VLAN mode - must be 'default' or 'trunk'",
		},

		"vlans": schema.StringAttribute{
			Optional:            true,
			Computed:            false,
			MarkdownDescription: "Comma-separated list of VLAN-IDs (required for VLAN mode trunk)",
		},
	}
}

func (r *MachinePoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: machinePoolDesc + "\n\n" +
			"~> Be aware that changing the version or the primary_disk_device will cause a new [Revision that will be applied immediately, causing a reboot of all Machines](https://meltcloud.io/docs/guides/machine-pools/upgrade.html#revisions).",

		Attributes: machinePoolResourceAttributes(),

		Blocks: map[string]schema.Block{
			"network_configuration": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: networkConfigurationResourceAttributes(),
				},
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

	var networkConfigurations []NetworkConfigurationResourceModel
	diags := data.NetworkConfigurations.ElementsAs(ctx, &networkConfigurations, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	machinePoolCreateInput := &client.MachinePoolCreateInput{
		Name:                       data.Name.ValueString(),
		PrimaryDiskDevice:          data.PrimaryDiskDevice.ValueString(),
		ReuseExistingRootPartition: data.ReuseExistingRootPartition.ValueBool(),
		UserVersion:                data.Version.ValueString(),
		NetworkConfigurations:      r.networkConfigurationInput(networkConfigurations),
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

func (r *MachinePoolResource) networkConfigurationInput(networkConfigurations []NetworkConfigurationResourceModel) []client.NetworkConfiguration {
	var networkConfigurationInput []client.NetworkConfiguration
	for _, networkConfiguration := range networkConfigurations {
		networkConfigurationInput = append(networkConfigurationInput, client.NetworkConfiguration{
			Type:       networkConfiguration.Type.ValueString(),
			Interfaces: networkConfiguration.Interfaces.ValueString(),
			VLANMode:   networkConfiguration.VLANMode.ValueString(),
			VLANs:      networkConfiguration.VLANs.ValueString(),
		})
	}
	return networkConfigurationInput
}

func (r *MachinePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MachinePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.MachinePool().Get(ctx, data.ClusterId.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine pool, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.MachinePool.Name)
	data.PrimaryDiskDevice = types.StringValue(result.MachinePool.PrimaryDiskDevice)
	data.ReuseExistingRootPartition = types.BoolValue(result.MachinePool.ReuseExistingRootPartition)
	data.Version = types.StringValue(result.MachinePool.UserVersion)
	data.PatchVersion = types.StringValue(result.MachinePool.PatchVersion)

	var networkConfigurations []NetworkConfigurationResourceModel
	for _, networkConfiguration := range result.MachinePool.NetworkConfigurations {
		networkConfigurations = append(networkConfigurations, NetworkConfigurationResourceModel{
			Type:       types.StringValue(networkConfiguration.Type),
			Interfaces: types.StringValue(networkConfiguration.Interfaces),
			VLANMode:   types.StringValue(networkConfiguration.VLANMode),
			VLANs:      types.StringValue(networkConfiguration.VLANs),
		})
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_configuration"), networkConfigurations)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachinePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MachinePoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var networkConfigurations []NetworkConfigurationResourceModel
	diags := data.NetworkConfigurations.ElementsAs(ctx, &networkConfigurations, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	machinePoolUpdateInput := &client.MachinePoolUpdateInput{
		Name:                       data.Name.ValueString(),
		PrimaryDiskDevice:          data.PrimaryDiskDevice.ValueString(),
		ReuseExistingRootPartition: data.ReuseExistingRootPartition.ValueBool(),
		UserVersion:                data.Version.ValueString(),
		NetworkConfigurations:      r.networkConfigurationInput(networkConfigurations),
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

var machinePoolImportIDPattern = regexp.MustCompile(`clusters/(\d+)/machine_pools/(\d+)`)

func (r *MachinePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := machinePoolImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 3 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", machinePoolImportIDPattern.String()))
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

	if resp.Diagnostics.HasError() {
		return
	}
}

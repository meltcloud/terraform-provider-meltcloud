package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
var _ resource.Resource = &MachineResource{}
var _ resource.ResourceWithImportState = &MachineResource{}

func NewMachineResource() resource.Resource {
	return &MachineResource{}
}

// MachineResource defines the resource implementation.
type MachineResource struct {
	client *client.Client
}

// MachineResourceModel describes the resource data model.
type MachineResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	UUID          types.String `tfsdk:"uuid"`
	Name          types.String `tfsdk:"name"`
	MachinePoolID types.Int64  `tfsdk:"machine_pool_id"`
	Labels        types.List   `tfsdk:"label"`
}

type LabelResourceModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (r *MachineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

const machineDesc string = "[Machines](https://meltcloud.io/docs/guides/machines/intro.html) are bare-metal or virtualized computers designated as worker nodes for the Kubernetes Clusters provided by the meltcloud platform."

func machineResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Machine in meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"uuid": schema.StringAttribute{
			MarkdownDescription: "UUID of the Machine",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Machine",
			Optional:            true,
		},
		"machine_pool_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the associated machine pool",
			Optional:            true,
		},
	}
}

func labelResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"key": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The key of the label, for example 'topology.kubernetes.io/zone'",
		},

		"value": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The value of the label, for example 'my-zone-1'",
		},
	}
}

func (r *MachineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: machineDesc + "\n\n" +
			"This resource [pre-registers](https://meltcloud.io/docs/guides/machines/intro.html#pre-register) Machines for a later boot.\n\n" +
			"~> Be aware that changing the name will cause a new [Revision that will be applied immediately, causing a reboot of the Machine](https://meltcloud.io/docs/guides/machines/intro.html#revisions).",

		Attributes: machineResourceAttributes(),

		Blocks: map[string]schema.Block{
			"label": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: labelResourceAttributes(),
				},
			},
		},
	}
}

func (r *MachineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uuid, err := uuid.Parse(data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("UUID invalid: %s", err))
		return
	}

	var labels []LabelResourceModel
	diags := data.Labels.ElementsAs(ctx, &labels, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	machineCreateInput := &client.MachineCreateInput{
		UUID:          uuid,
		Name:          data.Name.ValueString(),
		MachinePoolID: data.MachinePoolID.ValueInt64(),
		Labels:        r.labelInput(labels),
	}

	result, err2 := r.client.Machine().Create(ctx, machineCreateInput)
	if err2 != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create machine, got error: %s", err2))
		return
	}

	data.ID = types.Int64Value(result.Machine.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Machine().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine, got error: %s", err))
		return
	}

	data.UUID = types.StringValue(result.Machine.UUID.String())
	data.Name = types.StringValue(result.Machine.Name)
	data.MachinePoolID = types.Int64Value(result.Machine.MachinePoolID)

	var labels []LabelResourceModel
	for _, label := range result.Machine.Labels {
		labels = append(labels, LabelResourceModel{
			Key:   types.StringValue(label.Key),
			Value: types.StringValue(label.Value),
		})
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("label"), labels)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels []LabelResourceModel
	diags := data.Labels.ElementsAs(ctx, &labels, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	machineUpdateInput := &client.MachineUpdateInput{
		Name:          data.Name.ValueString(),
		MachinePoolID: data.MachinePoolID.ValueInt64(),
		Labels:        r.labelInput(labels),
	}

	result, err := r.client.Machine().Update(ctx, data.ID.ValueInt64(), machineUpdateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine, got error: %s", err))
			return
		}

		_, err := r.client.Machine().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update machine, got error: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Machine().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete machine, got error: %s", err))
		return
	}
}

var machineImportIDPattern = regexp.MustCompile(`machines/(\d+)`)

func (r *MachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := machineImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", machineImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *MachineResource) labelInput(labels []LabelResourceModel) []client.Label {
	var labelInput []client.Label
	for _, label := range labels {
		labelInput = append(labelInput, client.Label{
			Key:   label.Key.ValueString(),
			Value: label.Value.ValueString(),
		})
	}
	return labelInput
}

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NestInstanceResource{}
	_ resource.ResourceWithImportState = &NestInstanceResource{}
)

func NewNestInstanceResource() resource.Resource {
	return &NestInstanceResource{}
}

type NestInstanceResource struct {
	client *client.Client
}

type NestInstanceResourceModel struct {
	ID                      types.Int64  `tfsdk:"id"`
	NestID                  types.Int64  `tfsdk:"nest_id"`
	Name                    types.String `tfsdk:"name"`
	IP                      types.String `tfsdk:"ip"`
	ZoneID                  types.Int64  `tfsdk:"zone_id"`
	InstallDiskDevice       types.String `tfsdk:"install_disk_device"`
	InstallDiskMirror       types.Bool   `tfsdk:"install_disk_mirror"`
	InstallDiskMirrorDevice types.String `tfsdk:"install_disk_mirror_device"`
	EtcdDisk                types.Bool   `tfsdk:"etcd_disk"`
	EtcdDiskDevice          types.String `tfsdk:"etcd_disk_device"`
	EtcdDiskMirror          types.Bool   `tfsdk:"etcd_disk_mirror"`
	EtcdDiskMirrorDevice    types.String `tfsdk:"etcd_disk_mirror_device"`
	Status                  types.String `tfsdk:"status"`
	Architecture            types.String `tfsdk:"architecture"`
	Version                 types.String `tfsdk:"version"`
	URLISOAMD64             types.String `tfsdk:"url_iso_amd64"`
	URLISOARM64             types.String `tfsdk:"url_iso_arm64"`
}

func (r *NestInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_instance"
}

const nestInstanceDesc = "A [Nest Instance](https://docs.meltcloud.io/concepts/nest) is one machine of the Nest. " +
	"Adding one builds its install ISO: boot the machine from it, and it installs and joins the Nest. " +
	"Creating the resource waits until the ISO is built, not until the instance has joined; `status` turns from `installing` to `ready` once it has.\n\n" +
	"Needs the API Key of the admin Organization. A Nest has at most seven instances, and adds one at a time.\n\n" +
	"~> A Nest Instance cannot be removed yet, so destroying or replacing one fails."

func immutableString(description string, required bool) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Required:            required,
		Optional:            !required,
		Computed:            !required,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
	}
}

func immutableBool(description string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace(), boolplanmodifier.UseStateForUnknown()},
	}
}

func computedString(description string, sensitive bool) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: description, Computed: true, Sensitive: sensitive}
}

func nestInstanceResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Nest Instance on meltcloud",
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"nest_id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "ID of the Nest the instance belongs to",
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"name": immutableString("Name of the Nest Instance, unique within the Nest", true),
		"ip":   immutableString("The address the instance gets on the Nest's primary host network, from that network's IP Pool", true),
		"zone_id": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "ID of the [Zone](https://docs.meltcloud.io/concepts/regions-and-zones) the instance is in. Sets the Kubernetes labels `topology.kubernetes.io/region` and `topology.kubernetes.io/zone` on its node",
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
		},
		"install_disk_device":        immutableString("Device the instance installs to, for example `/dev/sda`. Leave it out to let the installer pick", false),
		"install_disk_mirror":        immutableBool("Mirror the install disk"),
		"install_disk_mirror_device": immutableString("Device the install disk is mirrored to", false),
		"etcd_disk":                  immutableBool("Keep etcd on a disk of its own"),
		"etcd_disk_device":           immutableString("Device etcd is kept on", false),
		"etcd_disk_mirror":           immutableBool("Mirror the etcd disk"),
		"etcd_disk_mirror_device":    immutableString("Device the etcd disk is mirrored to", false),
		"status":                     computedString("Status of the instance: `installing` until it has joined the Nest, then `ready`", false),
		"architecture":               computedString("CPU architecture the instance reported once it booted", false),
		"version":                    computedString("Nest version the instance runs", false),
		"url_iso_amd64":              computedString("URL of the install ISO for amd64 machines. It carries a download token", true),
		"url_iso_arm64":              computedString("URL of the install ISO for arm64 machines. It carries a download token", true),
	}
}

func (r *NestInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: nestInstanceDesc,
		Attributes:          nestInstanceResourceAttributes(),
	}
}

func (r *NestInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NestInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NestInstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Nest().CreateInstance(ctx, &client.NestInstanceCreateInput{
		Name:                    data.Name.ValueString(),
		IP:                      data.IP.ValueString(),
		ZoneID:                  data.ZoneID.ValueInt64(),
		InstallDiskDevice:       knownString(data.InstallDiskDevice),
		InstallDiskMirror:       knownBool(data.InstallDiskMirror),
		InstallDiskMirrorDevice: knownString(data.InstallDiskMirrorDevice),
		EtcdDisk:                knownBool(data.EtcdDisk),
		EtcdDiskDevice:          knownString(data.EtcdDiskDevice),
		EtcdDiskMirror:          knownBool(data.EtcdDiskMirror),
		EtcdDiskMirrorDevice:    knownString(data.EtcdDiskMirrorDevice),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add Nest Instance, got error: %s", err))
		return
	}

	// Saved first, so an instance whose ISO never builds is still in the state.
	setNestInstanceValues(result.NestInstance, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := r.awaitImage(ctx, result.NestInstance.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("The install ISO of Nest Instance %d was not built, got error: %s", result.NestInstance.ID, err))
		return
	}

	setNestInstanceValues(instance, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NestInstanceResource) awaitImage(ctx context.Context, id int64) (*client.NestInstance, *client.Error) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result, err := r.client.Nest().GetInstance(ctx, id)
			if err != nil {
				return nil, err
			}
			if image := result.NestInstance.Image; image != nil && image.Status == "ready" {
				return result.NestInstance, nil
			}
		case <-ctx.Done():
			return nil, &client.Error{Err: ctx.Err()}
		}
	}
}

func (r *NestInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NestInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Nest().GetInstance(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest Instance, got error: %s", err))
		return
	}

	setNestInstanceValues(result.NestInstance, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NestInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Supported", "A Nest Instance cannot be changed.")
}

func (r *NestInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Not Supported",
		"A Nest Instance cannot be removed yet. To stop managing it, remove it from the state with `terraform state rm`.")
}

var nestInstanceImportIDPattern = regexp.MustCompile(`nest_instances/(\d+)`)

func (r *NestInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := nestInstanceImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", nestInstanceImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func setNestInstanceValues(instance *client.NestInstance, data *NestInstanceResourceModel) {
	data.ID = types.Int64Value(instance.ID)
	data.NestID = types.Int64Value(instance.NestID)
	data.Name = types.StringValue(instance.Name)
	data.IP = types.StringValue(instance.IP)
	data.ZoneID = types.Int64Value(instance.ZoneID)
	data.InstallDiskDevice = types.StringValue(instance.InstallDiskDevice)
	data.InstallDiskMirror = types.BoolValue(instance.InstallDiskMirror)
	data.InstallDiskMirrorDevice = types.StringValue(instance.InstallDiskMirrorDevice)
	data.EtcdDisk = types.BoolValue(instance.EtcdDisk)
	data.EtcdDiskDevice = types.StringValue(instance.EtcdDiskDevice)
	data.EtcdDiskMirror = types.BoolValue(instance.EtcdDiskMirror)
	data.EtcdDiskMirrorDevice = types.StringValue(instance.EtcdDiskMirrorDevice)
	data.Status = types.StringValue(instance.Status)
	data.Architecture = types.StringValue(instance.Architecture)
	data.Version = types.StringValue(instance.Version)

	data.URLISOAMD64 = types.StringNull()
	data.URLISOARM64 = types.StringNull()
	if instance.Image != nil && instance.Image.Status == "ready" {
		data.URLISOAMD64 = types.StringValue(instance.Image.URLISOAMD64)
		data.URLISOARM64 = types.StringValue(instance.Image.URLISOARM64)
	}
}

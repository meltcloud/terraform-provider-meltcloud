package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &EnrollmentImageResource{}
var _ resource.ResourceWithImportState = &EnrollmentImageResource{}
var _ resource.ResourceWithValidateConfig = &EnrollmentImageResource{}

func NewEnrollmentImageResource() resource.Resource {
	return &EnrollmentImageResource{}
}

// EnrollmentImageResource defines the resource implementation.
type EnrollmentImageResource struct {
	client *client.Client
}

// EnrollmentImageResourceModel describes the resource data model.
type EnrollmentImageResourceModel struct {
	ID                        types.Int64       `tfsdk:"id"`
	Name                      types.String      `tfsdk:"name"`
	ExpiresAt                 timetypes.RFC3339 `tfsdk:"expires_at"`
	InstallDiskDevice         types.String      `tfsdk:"install_disk_device"`
	InstallDiskForceOverwrite types.Bool        `tfsdk:"install_disk_force_overwrite"`
	InstallDiskMirror         types.Bool        `tfsdk:"install_disk_mirror"`
	InstallDiskMirrorDevice   types.String      `tfsdk:"install_disk_mirror_device"`
	VLAN                      types.Int64       `tfsdk:"vlan"`
	EnableHTTP                types.Bool        `tfsdk:"enable_http"`
	HTTPURLISOAMD64           types.String      `tfsdk:"http_url_iso_amd64"`
	HTTPURLISOARM64           types.String      `tfsdk:"http_url_iso_arm64"`
	HTTPSURLISOAMD64          types.String      `tfsdk:"https_url_iso_arm64"`
	HTTPSURLISOARM64          types.String      `tfsdk:"https_url_iso_amd64"`
	LastUsedAt                timetypes.RFC3339 `tfsdk:"last_used_at"`
}

func (r *EnrollmentImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enrollment_image"
}

const enrollmentImageDesc string = "An [Enrollment Image](https://docs.meltcloud.io/tasks/enrollment-images) creates bootable images to enroll your Machines into your meltcloud organization."

func enrollmentImageResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Enrollment Image",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the Enrollment Image, not case-sensitive. Must be unique within the organization.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"expires_at": schema.StringAttribute{
			CustomType:          timetypes.RFC3339Type{},
			MarkdownDescription: "Timestamp when the image should expire",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"install_disk_device": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Device path of the disk where the system should be installed to. Use a `/dev/disk/by-path/` path (i.e. `/dev/disk/by-path/pci-0000:00:17.0-ata-1`) rather than a kernel name such as `/dev/sda`, as those can change between boots; see [Choosing Disk Devices](https://docs.meltcloud.io/tasks/enrollment-images#choosing-disk-devices). If not specified, the disk is auto-detected, which only works if Linux sees exactly one block device (i.e. a single attached disk, or a Dell BOSS RAID pair that shows up as one). It must be specified if `install_disk_mirror` is enabled.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"install_disk_force_overwrite": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Force overwrite disk if it contains unknown data",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"install_disk_mirror": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Whether the install disk is mirrored onto a second disk (RAID1) for redundancy. Requires `install_disk_device` and `install_disk_mirror_device` to be set, since auto-detection cannot decide which of two disks is the primary and which the mirror.",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"install_disk_mirror_device": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Device path of the disk used as the mirror, i.e. `/dev/disk/by-path/pci-0000:00:17.0-ata-2`. Required (and only allowed) if `install_disk_mirror` is enabled, and must differ from `install_disk_device`. It is never auto-detected.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"vlan": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The VLAN to use as the enrollment network",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"enable_http": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Whether the images should be downloadable via insecure HTTP",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"http_url_iso_amd64": schema.StringAttribute{
			MarkdownDescription: "URL to download the ISO for AMD64 via HTTP",
			Computed:            true,
			Sensitive:           true,
		},
		"http_url_iso_arm64": schema.StringAttribute{
			MarkdownDescription: "URL to download the ISO for ARM64 via HTTP",
			Computed:            true,
			Sensitive:           true,
		},
		"https_url_iso_amd64": schema.StringAttribute{
			MarkdownDescription: "URL to download the ISO for AMD64 via HTTPS",
			Computed:            true,
			Sensitive:           true,
		},
		"https_url_iso_arm64": schema.StringAttribute{
			MarkdownDescription: "URL to download the ISO for ARM64 via HTTPS",
			Computed:            true,
			Sensitive:           true,
		},
		"last_used_at": schema.StringAttribute{
			CustomType:          timetypes.RFC3339Type{},
			MarkdownDescription: "Timestamp when the image was last used for an enrollment",
			Computed:            true,
		},
	}
}

func (r *EnrollmentImageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: enrollmentImageDesc,

		Attributes: enrollmentImageResourceAttributes(),
	}
}

// ValidateConfig enforces the disk combinations the machine agent accepts: a mirrored install
// cannot auto-detect its disks, because auto-detection cannot decide which of two disks is the
// primary and which the copy.
func (r *EnrollmentImageResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data EnrollmentImageResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.InstallDiskMirror.IsUnknown() {
		return
	}

	if data.InstallDiskMirror.ValueBool() {
		if data.InstallDiskDevice.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("install_disk_device"),
				"Missing Attribute Configuration",
				"install_disk_device must be set if install_disk_mirror is enabled, as the disks cannot be auto-detected for a mirrored install.",
			)
		}

		if data.InstallDiskMirrorDevice.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("install_disk_mirror_device"),
				"Missing Attribute Configuration",
				"install_disk_mirror_device must be set if install_disk_mirror is enabled, as the disks cannot be auto-detected for a mirrored install.",
			)
		}

		if !data.InstallDiskDevice.IsUnknown() && !data.InstallDiskMirrorDevice.IsUnknown() &&
			!data.InstallDiskDevice.IsNull() && data.InstallDiskDevice.Equal(data.InstallDiskMirrorDevice) {
			resp.Diagnostics.AddAttributeError(
				path.Root("install_disk_mirror_device"),
				"Invalid Attribute Combination",
				"install_disk_mirror_device must differ from install_disk_device.",
			)
		}
	} else if !data.InstallDiskMirrorDevice.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("install_disk_mirror_device"),
			"Invalid Attribute Combination",
			"install_disk_mirror_device can only be set if install_disk_mirror is enabled.",
		)
	}
}

func (r *EnrollmentImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnrollmentImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EnrollmentImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	expiresAt, diagErr := data.ExpiresAt.ValueRFC3339Time()
	if diagErr != nil {
		resp.Diagnostics = diagErr
		return
	}

	var vlan *int64
	if !data.VLAN.IsNull() && !data.VLAN.IsUnknown() {
		vlan = data.VLAN.ValueInt64Pointer()
	}

	var installDiskForceOverwrite *bool
	if !data.InstallDiskForceOverwrite.IsNull() && !data.InstallDiskForceOverwrite.IsUnknown() {
		installDiskForceOverwrite = data.InstallDiskForceOverwrite.ValueBoolPointer()
	}

	var installDiskMirror *bool
	if !data.InstallDiskMirror.IsNull() && !data.InstallDiskMirror.IsUnknown() {
		installDiskMirror = data.InstallDiskMirror.ValueBoolPointer()
	}

	var installDiskDevice *string
	if !data.InstallDiskDevice.IsNull() && !data.InstallDiskDevice.IsUnknown() {
		installDiskDevice = data.InstallDiskDevice.ValueStringPointer()
	}

	var installDiskMirrorDevice *string
	if !data.InstallDiskMirrorDevice.IsNull() && !data.InstallDiskMirrorDevice.IsUnknown() {
		installDiskMirrorDevice = data.InstallDiskMirrorDevice.ValueStringPointer()
	}

	var enableHTTP *bool
	if !data.EnableHTTP.IsNull() && !data.EnableHTTP.IsUnknown() {
		enableHTTP = data.EnableHTTP.ValueBoolPointer()
	}

	enrollmentImageCreateInput := &client.EnrollmentImageCreateInput{
		Name:                      data.Name.ValueString(),
		ExpiresAt:                 expiresAt.UTC(),
		InstallDiskDevice:         installDiskDevice,
		InstallDiskForceOverwrite: installDiskForceOverwrite,
		InstallDiskMirror:         installDiskMirror,
		InstallDiskMirrorDevice:   installDiskMirrorDevice,
		VLAN:                      vlan,
		EnableHTTP:                enableHTTP,
	}

	result, err := r.client.EnrollmentImage().Create(ctx, enrollmentImageCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create enrollment image, got error: %s", err))
		return
	}
	if result.Operation == nil {
		resp.Diagnostics.AddError("Server Error", "Created enrollment image, but did not get operation")
		return
	}

	_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error during creation of enrollment image, got error: %s", err))
		return
	}

	result, err = r.client.EnrollmentImage().Get(ctx, result.EnrollmentImage.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read enrollment image, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.EnrollmentImage.ID)
	r.setValues(result.EnrollmentImage, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnrollmentImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnrollmentImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.EnrollmentImage().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read enrollment image, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.EnrollmentImage.Name)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(result.EnrollmentImage.ExpiresAt.UTC())
	r.setValues(result.EnrollmentImage, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnrollmentImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource Update Not Implemented", "enrollment_image does not support updates")
}

func (r *EnrollmentImageResource) setValues(result *client.EnrollmentImage, data *EnrollmentImageResourceModel) {
	data.VLAN = types.Int64PointerValue(result.VLAN)
	data.EnableHTTP = types.BoolValue(result.EnableHTTP)
	data.InstallDiskDevice = types.StringPointerValue(result.InstallDiskDevice)
	data.InstallDiskForceOverwrite = types.BoolValue(result.InstallDiskForceOverwrite)
	data.InstallDiskMirror = types.BoolValue(result.InstallDiskMirror)
	data.InstallDiskMirrorDevice = types.StringPointerValue(result.InstallDiskMirrorDevice)
	data.HTTPURLISOAMD64 = types.StringValue(result.HTTPURLISOAMD64)
	data.HTTPURLISOARM64 = types.StringValue(result.HTTPURLISOARM64)
	data.HTTPSURLISOAMD64 = types.StringValue(result.HTTPSURLISOAMD64)
	data.HTTPSURLISOARM64 = types.StringValue(result.HTTPSURLISOARM64)
	data.LastUsedAt = timetypes.NewRFC3339TimePointerValue(result.LastUsedAt)
}

func (r *EnrollmentImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EnrollmentImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.EnrollmentImage().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete enrollment image, got error: %s", err))
		return
	}
}

var enrollmentImageImportIDPattern = regexp.MustCompile(`enrollment_images/(\d+)`)

func (r *EnrollmentImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := enrollmentImageImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", enrollmentImageImportIDPattern.String()))
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

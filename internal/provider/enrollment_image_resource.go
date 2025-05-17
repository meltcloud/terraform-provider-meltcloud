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
	VLAN                      types.Int64       `tfsdk:"vlan"`
	EnableHTTP                types.Bool        `tfsdk:"enable_http"`
	HTTPURLISOAMD64           types.String      `tfsdk:"http_url_iso_amd64"`
	HTTPURLISOARM64           types.String      `tfsdk:"http_url_iso_arm64"`
	HTTPSURLISOAMD64          types.String      `tfsdk:"https_url_iso_arm64"`
	HTTPSURLISOARM64          types.String      `tfsdk:"https_url_iso_amd64"`
	IPXEScriptHTTPAMD64       types.String      `tfsdk:"ipxe_script_http_amd64"`
	IPXEScriptHTTPARM64       types.String      `tfsdk:"ipxe_script_http_arm64"`
	IPXEScriptHTTPSAMD64      types.String      `tfsdk:"ipxe_script_https_amd64"`
	IPXEScriptHTTPSARM64      types.String      `tfsdk:"ipxe_script_https_arm64"`
	LastUsedAt                timetypes.RFC3339 `tfsdk:"last_used_at"`
}

func (r *EnrollmentImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enrollment_image"
}

const enrollmentImageDesc string = "An [Enrollment Image](https://meltcloud.io/docs/guides/boot-config/enrollment-image.html) creates bootable images to enroll your Machines into your meltcloud organization."

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
			Required:            true,
			MarkdownDescription: "Device path (i.e. /dev/vda) of the disk where system should be installed to",
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
		"ipxe_script_http_amd64": schema.StringAttribute{
			MarkdownDescription: "iPXE script to boot the ISO for AMD64 via HTTP",
			Computed:            true,
			Sensitive:           true,
		},
		"ipxe_script_http_arm64": schema.StringAttribute{
			MarkdownDescription: "iPXE script to boot the ISO for ARM64 via HTTP",
			Computed:            true,
			Sensitive:           true,
		},
		"ipxe_script_https_amd64": schema.StringAttribute{
			MarkdownDescription: "iPXE script to boot the ISO for AMD64 via HTTPS",
			Computed:            true,
			Sensitive:           true,
		},
		"ipxe_script_https_arm64": schema.StringAttribute{
			MarkdownDescription: "iPXE script to boot the ISO for ARM64 via HTTPS",
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

	var enableHTTP *bool
	if !data.EnableHTTP.IsNull() && !data.EnableHTTP.IsUnknown() {
		enableHTTP = data.EnableHTTP.ValueBoolPointer()
	}

	enrollmentImageCreateInput := &client.EnrollmentImageCreateInput{
		Name:                      data.Name.ValueString(),
		ExpiresAt:                 expiresAt.UTC(),
		InstallDiskDevice:         data.InstallDiskDevice.ValueString(),
		InstallDiskForceOverwrite: installDiskForceOverwrite,
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
	data.InstallDiskDevice = types.StringValue(result.EnrollmentImage.InstallDiskDevice)
	r.setValues(result.EnrollmentImage, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnrollmentImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource Update Not Implemented", "enrollment_image does not support updates")
}

func (r *EnrollmentImageResource) setValues(result *client.EnrollmentImage, data *EnrollmentImageResourceModel) {
	data.VLAN = types.Int64PointerValue(result.VLAN)
	data.EnableHTTP = types.BoolValue(result.EnableHTTP)
	data.InstallDiskForceOverwrite = types.BoolValue(result.InstallDiskForceOverwrite)
	data.HTTPURLISOAMD64 = types.StringValue(result.HTTPURLISOAMD64)
	data.HTTPURLISOARM64 = types.StringValue(result.HTTPURLISOARM64)
	data.HTTPSURLISOAMD64 = types.StringValue(result.HTTPSURLISOAMD64)
	data.HTTPSURLISOARM64 = types.StringValue(result.HTTPSURLISOARM64)
	data.IPXEScriptHTTPAMD64 = types.StringValue(result.IPXEScriptHTTPAMD64)
	data.IPXEScriptHTTPARM64 = types.StringValue(result.IPXEScriptHTTPARM64)
	data.IPXEScriptHTTPSAMD64 = types.StringValue(result.IPXEScriptHTTPSAMD64)
	data.IPXEScriptHTTPSARM64 = types.StringValue(result.IPXEScriptHTTPSARM64)
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

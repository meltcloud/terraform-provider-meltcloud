package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NestInstanceImageResource{}
	_ resource.ResourceWithImportState = &NestInstanceImageResource{}
)

func NewNestInstanceImageResource() resource.Resource {
	return &NestInstanceImageResource{}
}

type NestInstanceImageResource struct {
	client *client.Client
}

type NestInstanceImageResourceModel struct {
	ID               types.Int64       `tfsdk:"id"`
	NestInstanceID   types.Int64       `tfsdk:"nest_instance_id"`
	Name             types.String      `tfsdk:"name"`
	ExpiresAt        timetypes.RFC3339 `tfsdk:"expires_at"`
	Status           types.String      `tfsdk:"status"`
	LastUsedAt       timetypes.RFC3339 `tfsdk:"last_used_at"`
	HTTPSURLISOAMD64 types.String      `tfsdk:"https_url_iso_amd64"`
	HTTPSURLISOARM64 types.String      `tfsdk:"https_url_iso_arm64"`
	HTTPURLISOAMD64  types.String      `tfsdk:"http_url_iso_amd64"`
	HTTPURLISOARM64  types.String      `tfsdk:"http_url_iso_arm64"`

	InstallDiskForceOverwrite types.Bool `tfsdk:"install_disk_force_overwrite"`
	EnableHTTP                types.Bool `tfsdk:"enable_http"`
}

func (r *NestInstanceImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_instance_image"
}

const nestInstanceImageDesc = "An [Installer Image](https://docs.meltcloud.io/tasks/nest/create-instance#create-an-installer-image) of a Nest Instance. " +
	"Boot the host from it to install the instance and join it to the Nest.\n\n" +
	"Needs the API Key of the admin Organization."

const nestInstanceImageForceOverwriteDesc = "Force overwrite disk if it contains any data"

func (r *NestInstanceImageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: nestInstanceImageDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Internal ID of the image on meltcloud",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"nest_instance_id": immutableInt64("ID of the Nest Instance the image installs", true),
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the image, unique within the Nest Instance",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"expires_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				Required:            true,
				MarkdownDescription: "Timestamp after which the image can no longer be used for installation",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the image: `pending` while it builds, then `ready`, or `failed` if the build failed",
			},
			"install_disk_force_overwrite": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: nestInstanceImageForceOverwriteDesc,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"enable_http": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the image can be downloaded via insecure HTTP",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"last_used_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				Computed:            true,
				MarkdownDescription: "Timestamp when an install last used the image",
			},
			"https_url_iso_amd64": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "URL to download the ISO for amd64 via HTTPS",
			},
			"https_url_iso_arm64": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "URL to download the ISO for arm64 via HTTPS",
			},
			"http_url_iso_amd64": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "URL to download the ISO for amd64 via HTTP, if `enable_http` is set",
			},
			"http_url_iso_arm64": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "URL to download the ISO for arm64 via HTTP, if `enable_http` is set",
			},
		},
	}
}

func (r *NestInstanceImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NestInstanceImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NestInstanceImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	expiresAt, diags := data.ExpiresAt.ValueRFC3339Time()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := data.NestInstanceID.ValueInt64()
	result, err := r.client.Nest().CreateInstanceImage(ctx, instanceID, &client.NestInstanceImageCreateInput{
		Name:                      data.Name.ValueString(),
		ExpiresAt:                 expiresAt.UTC(),
		InstallDiskForceOverwrite: data.InstallDiskForceOverwrite.ValueBool(),
		EnableHTTP:                data.EnableHTTP.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Nest Instance image, got error: %s", err))
		return
	}

	// Saved first, so an image whose build fails is still in the state.
	setNestInstanceImageValues(result.NestInstanceImage, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if result.Operation == nil {
		resp.Diagnostics.AddError("Server Error", "Created the Nest Instance image, but did not get its build operation")
		return
	}
	if _, err := r.client.Operation().PollUntilDone(ctx, result.Operation.ID); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to build Nest Instance image, got error: %s", err))
		return
	}

	result, err = r.client.Nest().GetInstanceImage(ctx, instanceID, result.NestInstanceImage.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest Instance image, got error: %s", err))
		return
	}

	setNestInstanceImageValues(result.NestInstanceImage, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NestInstanceImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NestInstanceImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Nest().GetInstanceImage(ctx, data.NestInstanceID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest Instance image, got error: %s", err))
		return
	}

	setNestInstanceImageValues(result.NestInstanceImage, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NestInstanceImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Supported", "A Nest Instance image cannot be changed.")
}

func (r *NestInstanceImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NestInstanceImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Nest().DeleteInstanceImage(ctx, data.NestInstanceID.ValueInt64(), data.ID.ValueInt64())
	if err != nil && err.HTTPStatusCode != 404 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Nest Instance image, got error: %s", err))
	}
}

var nestInstanceImageImportIDPattern = regexp.MustCompile(`nest_instances/(\d+)/images/(\d+)`)

func (r *NestInstanceImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := nestInstanceImageImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 3 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", nestInstanceImageImportIDPattern.String()))
		return
	}

	instanceID, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}
	id, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("nest_instance_id"), instanceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func setNestInstanceImageValues(image *client.NestInstanceImage, data *NestInstanceImageResourceModel) {
	data.ID = types.Int64Value(image.ID)
	data.NestInstanceID = types.Int64Value(image.NestInstanceID)
	data.Name = types.StringValue(image.Name)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(image.ExpiresAt.UTC())
	data.Status = types.StringValue(image.Status)
	data.LastUsedAt = timetypes.NewRFC3339TimePointerValue(image.LastUsedAt)
	data.HTTPSURLISOAMD64 = types.StringPointerValue(image.HTTPSURLISOAMD64)
	data.HTTPSURLISOARM64 = types.StringPointerValue(image.HTTPSURLISOARM64)
	data.HTTPURLISOAMD64 = types.StringPointerValue(image.HTTPURLISOAMD64)
	data.HTTPURLISOARM64 = types.StringPointerValue(image.HTTPURLISOARM64)
	data.InstallDiskForceOverwrite = types.BoolValue(image.InstallDiskForceOverwrite)
	data.EnableHTTP = types.BoolValue(image.EnableHTTP)
}

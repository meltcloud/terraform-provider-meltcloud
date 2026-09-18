package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &EnrollmentImageDataSource{}

func NewEnrollmentImageDataSource() datasource.DataSource {
	return &EnrollmentImageDataSource{}
}

// EnrollmentImageDataSource defines the data source implementation.
type EnrollmentImageDataSource struct {
	client *client.Client
}

type EnrollmentImageDataSourceModel struct {
	ID                        types.Int64       `tfsdk:"id"`
	Name                      types.String      `tfsdk:"name"`
	Status                    types.String      `tfsdk:"status"`
	ExpiresAt                 timetypes.RFC3339 `tfsdk:"expires_at"`
	NetworkProfileID          types.Int64       `tfsdk:"network_profile_id"`
	InstallDiskDevice         types.String      `tfsdk:"install_disk_device"`
	InstallDiskForceOverwrite types.Bool        `tfsdk:"install_disk_force_overwrite"`
	InstallDiskMirror         types.Bool        `tfsdk:"install_disk_mirror"`
	InstallDiskMirrorDevice   types.String      `tfsdk:"install_disk_mirror_device"`
	EnableHTTP                types.Bool        `tfsdk:"enable_http"`
	HTTPURLISOAMD64           types.String      `tfsdk:"http_url_iso_amd64"`
	HTTPURLISOARM64           types.String      `tfsdk:"http_url_iso_arm64"`
	HTTPSURLISOAMD64          types.String      `tfsdk:"https_url_iso_arm64"`
	HTTPSURLISOARM64          types.String      `tfsdk:"https_url_iso_amd64"`
	LastUsedAt                timetypes.RFC3339 `tfsdk:"last_used_at"`
}

func (d *EnrollmentImageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enrollment_image"
}

func (d *EnrollmentImageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: enrollmentImageDesc,

		Attributes: withLookupAttributes(map[string]schema.Attribute{
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the Enrollment Image",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: enrollmentImageResourceAttributes()["expires_at"].GetMarkdownDescription(),
				Computed:            true,
			},
			"network_profile_id": schema.Int64Attribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["network_profile_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"install_disk_device": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["install_disk_device"].GetMarkdownDescription(),
				Computed:            true,
			},
			"install_disk_force_overwrite": schema.BoolAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["install_disk_force_overwrite"].GetMarkdownDescription(),
				Computed:            true,
			},
			"install_disk_mirror": schema.BoolAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["install_disk_mirror"].GetMarkdownDescription(),
				Computed:            true,
			},
			"install_disk_mirror_device": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["install_disk_mirror_device"].GetMarkdownDescription(),
				Computed:            true,
			},
			"enable_http": schema.BoolAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["enable_http"].GetMarkdownDescription(),
				Computed:            true,
			},
			"http_url_iso_amd64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["http_url_iso_amd64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"http_url_iso_arm64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["http_url_iso_arm64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"https_url_iso_amd64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["https_url_iso_amd64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"https_url_iso_arm64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["https_url_iso_arm64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"last_used_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: enrollmentImageResourceAttributes()["last_used_at"].GetMarkdownDescription(),
				Computed:            true,
			},
		}, enrollmentImageResourceAttributes()["id"].GetMarkdownDescription(), enrollmentImageResourceAttributes()["name"].GetMarkdownDescription()),
	}
}

func (d *EnrollmentImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *EnrollmentImageDataSource) readEnrollmentImage(ctx context.Context, data EnrollmentImageDataSourceModel) (*client.EnrollmentImageResult, *client.Error) {
	if data.ID.IsNull() {
		return d.client.EnrollmentImage().GetByName(ctx, data.Name.ValueString())
	}
	return d.client.EnrollmentImage().Get(ctx, data.ID.ValueInt64())
}

func (d *EnrollmentImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnrollmentImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.readEnrollmentImage(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read enrollment image, got error: %s", err))
		return
	}
	enrollmentImage := result.EnrollmentImage

	data.ID = types.Int64Value(enrollmentImage.ID)
	data.Name = types.StringValue(enrollmentImage.Name)
	data.Status = types.StringValue(enrollmentImage.Status)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(enrollmentImage.ExpiresAt)
	data.NetworkProfileID = types.Int64Value(enrollmentImage.NetworkProfileID)
	data.InstallDiskDevice = types.StringPointerValue(enrollmentImage.InstallDiskDevice)
	data.EnableHTTP = types.BoolValue(enrollmentImage.EnableHTTP)
	data.InstallDiskForceOverwrite = types.BoolValue(enrollmentImage.InstallDiskForceOverwrite)
	data.InstallDiskMirror = types.BoolValue(enrollmentImage.InstallDiskMirror)
	data.InstallDiskMirrorDevice = types.StringPointerValue(enrollmentImage.InstallDiskMirrorDevice)
	data.HTTPURLISOAMD64 = types.StringValue(enrollmentImage.HTTPURLISOAMD64)
	data.HTTPURLISOARM64 = types.StringValue(enrollmentImage.HTTPURLISOARM64)
	data.HTTPSURLISOAMD64 = types.StringValue(enrollmentImage.HTTPSURLISOAMD64)
	data.HTTPSURLISOARM64 = types.StringValue(enrollmentImage.HTTPSURLISOARM64)
	data.LastUsedAt = timetypes.NewRFC3339TimePointerValue(enrollmentImage.LastUsedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

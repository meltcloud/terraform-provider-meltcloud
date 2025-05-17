package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-meltcloud/internal/client"
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

func (d *EnrollmentImageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enrollment_image"
}

func (d *EnrollmentImageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: enrollmentImageDesc,

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["id"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["name"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the Enrollment Image",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: enrollmentImageResourceAttributes()["expires_at"].GetMarkdownDescription(),
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
			"vlan": schema.Int64Attribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["vlan"].GetMarkdownDescription(),
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
			"ipxe_script_http_amd64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["ipxe_script_http_amd64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"ipxe_script_http_arm64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["ipxe_script_http_arm64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"ipxe_script_https_amd64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["ipxe_script_https_amd64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"ipxe_script_https_arm64": schema.StringAttribute{
				MarkdownDescription: enrollmentImageResourceAttributes()["ipxe_script_https_arm64"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"last_used_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: enrollmentImageResourceAttributes()["last_used_at"].GetMarkdownDescription(),
				Computed:            true,
			},
		},
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

func (d *EnrollmentImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnrollmentImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var enrollmentImage *client.EnrollmentImage
	if data.ID.ValueInt64() != 0 {
		result, err := d.client.EnrollmentImage().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read enrollment image by ID %d, got error: %s", data.ID.ValueInt64(), err))
			return
		}
		enrollmentImage = result.EnrollmentImage
	} else {
		result, err := d.client.EnrollmentImage().List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read enrollment images, got error: %s", err))
			return
		}

		for _, a := range result.EnrollmentImages {
			if strings.EqualFold(data.Name.ValueString(), a.Name) {
				enrollmentImage = a
				break
			}
		}

		if enrollmentImage == nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not find enrollment image by name %s", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.Int64Value(enrollmentImage.ID)
	data.Name = types.StringValue(enrollmentImage.Name)
	data.Status = types.StringValue(enrollmentImage.Status)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(enrollmentImage.ExpiresAt)
	data.InstallDiskDevice = types.StringValue(enrollmentImage.InstallDiskDevice)
	data.VLAN = types.Int64PointerValue(enrollmentImage.VLAN)
	data.EnableHTTP = types.BoolValue(enrollmentImage.EnableHTTP)
	data.InstallDiskForceOverwrite = types.BoolValue(enrollmentImage.InstallDiskForceOverwrite)
	data.HTTPURLISOAMD64 = types.StringValue(enrollmentImage.HTTPURLISOAMD64)
	data.HTTPURLISOARM64 = types.StringValue(enrollmentImage.HTTPURLISOARM64)
	data.HTTPSURLISOAMD64 = types.StringValue(enrollmentImage.HTTPSURLISOAMD64)
	data.HTTPSURLISOARM64 = types.StringValue(enrollmentImage.HTTPSURLISOARM64)
	data.IPXEScriptHTTPAMD64 = types.StringValue(enrollmentImage.IPXEScriptHTTPAMD64)
	data.IPXEScriptHTTPARM64 = types.StringValue(enrollmentImage.IPXEScriptHTTPARM64)
	data.IPXEScriptHTTPSAMD64 = types.StringValue(enrollmentImage.IPXEScriptHTTPSAMD64)
	data.IPXEScriptHTTPSARM64 = types.StringValue(enrollmentImage.IPXEScriptHTTPSARM64)
	data.LastUsedAt = timetypes.NewRFC3339TimePointerValue(enrollmentImage.LastUsedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

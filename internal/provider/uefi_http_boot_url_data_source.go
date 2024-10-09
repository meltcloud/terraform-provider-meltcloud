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
var _ datasource.DataSource = &UEFIHTTPBootURLDataSource{}

func NewUEFIHTTPBootURLDataSource() datasource.DataSource {
	return &UEFIHTTPBootURLDataSource{}
}

// UEFIHTTPBootURLDataSource defines the data source implementation.
type UEFIHTTPBootURLDataSource struct {
	client *client.Client
}

type UEFIHTTPBootURLDataSourceModel struct {
	ID                 types.Int64 `tfsdk:"id"`
	IPXEBootArtifactID types.Int64 `tfsdk:"ipxe_boot_artifact_id"`

	Name      types.String      `tfsdk:"name"`
	ExpiresAt timetypes.RFC3339 `tfsdk:"expires_at"`
	Protocols types.String      `tfsdk:"protocols"`
	HTTPURL   types.String      `tfsdk:"http_url"`
	HTTPSURL  types.String      `tfsdk:"https_url"`
}

func (d *UEFIHTTPBootURLDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uefi_http_boot_url"
}

func (d *UEFIHTTPBootURLDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: uefiHTTPBootURLDesc,

		Attributes: map[string]schema.Attribute{
			"ipxe_boot_artifact_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: uefiHTTPBootURLResourceAttributes()["ipxe_boot_artifact_id"].GetMarkdownDescription(),
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: uefiHTTPBootURLResourceAttributes()["id"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: uefiHTTPBootURLResourceAttributes()["name"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"protocols": schema.StringAttribute{
				MarkdownDescription: uefiHTTPBootURLResourceAttributes()["protocols"].GetMarkdownDescription(),
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: uefiHTTPBootURLResourceAttributes()["expires_at"].GetMarkdownDescription(),
				Computed:            true,
			},
			"http_url": schema.StringAttribute{
				MarkdownDescription: uefiHTTPBootURLResourceAttributes()["http_url"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"https_url": schema.StringAttribute{
				MarkdownDescription: uefiHTTPBootURLResourceAttributes()["https_url"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *UEFIHTTPBootURLDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UEFIHTTPBootURLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UEFIHTTPBootURLDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var uefiHTTPBootURL *client.UEFIHTTPBootURL
	if data.ID.ValueInt64() != 0 {
		result, err := d.client.UEFIHTTPBootURL().Get(ctx, data.IPXEBootArtifactID.ValueInt64(), data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read uefi http boot url by ID %d, got error: %s", data.ID.ValueInt64(), err))
			return
		}
		uefiHTTPBootURL = result.UEFIHTTPBootURL
	} else {
		result, err := d.client.UEFIHTTPBootURL().List(ctx, data.IPXEBootArtifactID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read uefi http boot urls, got error: %s", err))
			return
		}

		for _, u := range result.UEFIHTTPBootURLs {
			if strings.EqualFold(data.Name.ValueString(), u.Name) {
				uefiHTTPBootURL = u
				break
			}
		}

		if uefiHTTPBootURL == nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not find uefi http boot url by name %s", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.Int64Value(uefiHTTPBootURL.ID)
	data.Name = types.StringValue(uefiHTTPBootURL.Name)
	data.Protocols = types.StringValue(uefiHTTPBootURL.Protocols)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(uefiHTTPBootURL.ExpiresAt)
	data.HTTPURL = types.StringValue(uefiHTTPBootURL.HTTPURL)
	data.HTTPSURL = types.StringValue(uefiHTTPBootURL.HTTPSURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

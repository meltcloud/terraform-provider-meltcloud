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
var _ datasource.DataSource = &iPXEBootArtifactDataSource{}

func NewiPXEBootArtifactDataSource() datasource.DataSource {
	return &iPXEBootArtifactDataSource{}
}

// iPXEBootArtifactDataSource defines the data source implementation.
type iPXEBootArtifactDataSource struct {
	client *client.Client
}

type iPXEBootArtifactDataSourceModel struct {
	ID             types.Int64       `tfsdk:"id"`
	Name           types.String      `tfsdk:"name"`
	Status         types.String      `tfsdk:"status"`
	ExpiresAt      timetypes.RFC3339 `tfsdk:"expires_at"`
	DownloadURLISO types.String      `tfsdk:"download_url_iso"`
	DownloadURLPXE types.String      `tfsdk:"download_url_pxe"`
	DownloadURLEFI types.String      `tfsdk:"download_url_efi"`
}

func (d *iPXEBootArtifactDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipxe_boot_artifact"
}

func (d *iPXEBootArtifactDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: iPXEBootArtifactDesc,

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: iPXEBootArtifactResourceAttributes()["id"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: iPXEBootArtifactResourceAttributes()["name"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the iPXE Boot Artifact",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: iPXEBootArtifactResourceAttributes()["expires_at"].GetMarkdownDescription(),
				Computed:            true,
			},
			"download_url_iso": schema.StringAttribute{
				MarkdownDescription: iPXEBootArtifactResourceAttributes()["download_url_iso"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"download_url_pxe": schema.StringAttribute{
				MarkdownDescription: iPXEBootArtifactResourceAttributes()["download_url_pxe"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"download_url_efi": schema.StringAttribute{
				MarkdownDescription: iPXEBootArtifactResourceAttributes()["download_url_efi"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *iPXEBootArtifactDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *iPXEBootArtifactDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data iPXEBootArtifactDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var iPXEBootArtifact *client.IPXEBootArtifact
	if data.ID.ValueInt64() != 0 {
		result, err := d.client.IPXEBootArtifact().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe boot artifact by ID %d, got error: %s", data.ID.ValueInt64(), err))
			return
		}
		iPXEBootArtifact = result.IPXEBootArtifact
	} else {
		result, err := d.client.IPXEBootArtifact().List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe boot artifacts, got error: %s", err))
			return
		}

		for _, a := range result.IPXEBootArtifacts {
			if strings.EqualFold(data.Name.ValueString(), a.Name) {
				iPXEBootArtifact = a
				break
			}
		}

		if iPXEBootArtifact == nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not find ipxe boot artifact by name %s", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.Int64Value(iPXEBootArtifact.ID)
	data.Name = types.StringValue(iPXEBootArtifact.Name)
	data.Status = types.StringValue(iPXEBootArtifact.Status)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(iPXEBootArtifact.ExpiresAt)
	data.DownloadURLISO = types.StringValue(iPXEBootArtifact.DownloadURLISO)
	data.DownloadURLPXE = types.StringValue(iPXEBootArtifact.DownloadURLPXE)
	data.DownloadURLEFI = types.StringValue(iPXEBootArtifact.DownloadURLEFI)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

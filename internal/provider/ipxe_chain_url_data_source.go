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
var _ datasource.DataSource = &IPXEChainURLDataSource{}

func NewIPXEChainURLDataSource() datasource.DataSource {
	return &IPXEChainURLDataSource{}
}

// IPXEChainURLDataSource defines the data source implementation.
type IPXEChainURLDataSource struct {
	client *client.Client
}

type IPXEChainURLDataSourceModel struct {
	ID        types.Int64       `tfsdk:"id"`
	Name      types.String      `tfsdk:"name"`
	ExpiresAt timetypes.RFC3339 `tfsdk:"expires_at"`
	URL       types.String      `tfsdk:"url"`
	Script    types.String      `tfsdk:"script"`
}

func (d *IPXEChainURLDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipxe_chain_url"
}

func (d *IPXEChainURLDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: iPXEChainURLDesc,

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: iPXEChainURLResourceAttributes()["id"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: iPXEChainURLResourceAttributes()["name"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"expires_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: iPXEChainURLResourceAttributes()["expires_at"].GetMarkdownDescription(),
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: iPXEChainURLResourceAttributes()["url"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
			"script": schema.StringAttribute{
				MarkdownDescription: iPXEChainURLResourceAttributes()["script"].GetMarkdownDescription(),
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *IPXEChainURLDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IPXEChainURLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPXEChainURLDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var iPXEChainURL *client.IPXEChainURL
	if data.ID.ValueInt64() != 0 {
		result, err := d.client.IPXEChainURL().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe boot artifact by ID %d, got error: %s", data.ID.ValueInt64(), err))
			return
		}
		iPXEChainURL = result.IPXEChainURL
	} else {
		result, err := d.client.IPXEChainURL().List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe boot artifacts, got error: %s", err))
			return
		}

		for _, u := range result.IPXEChainURLs {
			if strings.EqualFold(data.Name.ValueString(), u.Name) {
				iPXEChainURL = u
				break
			}
		}

		if iPXEChainURL == nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not find ipxe boot artifact by name %s", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.Int64Value(iPXEChainURL.ID)
	data.Name = types.StringValue(iPXEChainURL.Name)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(iPXEChainURL.ExpiresAt)
	data.URL = types.StringValue(iPXEChainURL.URL)
	data.Script = types.StringValue(iPXEChainURL.Script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

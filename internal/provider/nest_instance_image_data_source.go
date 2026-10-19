package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &NestInstanceImageDataSource{}

func NewNestInstanceImageDataSource() datasource.DataSource {
	return &NestInstanceImageDataSource{}
}

type NestInstanceImageDataSource struct {
	client *client.Client
}

func (d *NestInstanceImageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_instance_image"
}

func (d *NestInstanceImageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	computed := func(description string, sensitive bool) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: description}
	}
	timestamp := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, CustomType: timetypes.RFC3339Type{}, MarkdownDescription: description}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "An [Installer Image](https://docs.meltcloud.io/tasks/nest/create-instance#create-an-installer-image) of a Nest Instance. Needs the API Key of the admin Organization.",
		Attributes: withLookupAttributes(map[string]schema.Attribute{
			"nest_instance_id":    schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the Nest Instance the image installs"},
			"expires_at":          timestamp("Timestamp after which the image can no longer be used for installation"),
			"status":              computed("Status of the image: `pending` while it builds, then `ready`", false),
			"last_used_at":        timestamp("Timestamp when an install last used the image"),
			"https_url_iso_amd64": computed("URL to download the ISO for amd64 via HTTPS", true),
			"https_url_iso_arm64": computed("URL to download the ISO for arm64 via HTTPS", true),
		}, "Internal ID of the image on meltcloud", "Name of the image, unique within the Nest Instance"),
	}
}

func (d *NestInstanceImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *NestInstanceImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NestInstanceImageResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *client.NestInstanceImageResult
	var err *client.Error
	instanceID := data.NestInstanceID.ValueInt64()
	if data.ID.IsNull() {
		result, err = d.client.Nest().GetInstanceImageByName(ctx, instanceID, data.Name.ValueString())
	} else {
		result, err = d.client.Nest().GetInstanceImage(ctx, instanceID, data.ID.ValueInt64())
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest Instance image, got error: %s", err))
		return
	}

	setNestInstanceImageValues(result.NestInstanceImage, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

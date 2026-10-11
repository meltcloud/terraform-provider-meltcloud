package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ datasource.DataSource = &NestInstanceDataSource{}

func NewNestInstanceDataSource() datasource.DataSource {
	return &NestInstanceDataSource{}
}

type NestInstanceDataSource struct {
	client *client.Client
}

func (d *NestInstanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_instance"
}

// Every attribute of the resource, read-only, except the lookup pair.
func (d *NestInstanceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceAttributes := nestInstanceResourceAttributes()
	attributes := map[string]schema.Attribute{}
	for name, attribute := range resourceAttributes {
		if name == "id" || name == "name" {
			continue
		}
		description := attribute.GetMarkdownDescription()
		switch attribute.(type) {
		case resourceschema.Int64Attribute:
			attributes[name] = schema.Int64Attribute{Computed: true, MarkdownDescription: description}
		case resourceschema.BoolAttribute:
			attributes[name] = schema.BoolAttribute{Computed: true, MarkdownDescription: description}
		default:
			attributes[name] = schema.StringAttribute{Computed: true, MarkdownDescription: description, Sensitive: attribute.IsSensitive()}
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: nestInstanceDesc,
		Attributes: withLookupAttributes(attributes,
			resourceAttributes["id"].GetMarkdownDescription(), resourceAttributes["name"].GetMarkdownDescription()),
	}
}

func (d *NestInstanceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *NestInstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NestInstanceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *client.NestInstanceResult
	var err *client.Error
	if data.ID.IsNull() {
		result, err = d.client.Nest().GetInstanceByName(ctx, data.Name.ValueString())
	} else {
		result, err = d.client.Nest().GetInstance(ctx, data.ID.ValueInt64())
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest Instance, got error: %s", err))
		return
	}

	setNestInstanceValues(result.NestInstance, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

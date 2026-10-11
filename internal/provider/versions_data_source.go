package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &FoundryVersionsDataSource{}
	_ datasource.DataSource = &NestVersionsDataSource{}
)

func NewFoundryVersionsDataSource() datasource.DataSource {
	return &FoundryVersionsDataSource{}
}

func NewNestVersionsDataSource() datasource.DataSource {
	return &NestVersionsDataSource{}
}

type FoundryVersionsDataSource struct {
	client *client.Client
}

type NestVersionsDataSource struct {
	client *client.Client
}

type FoundryVersionsDataSourceModel struct {
	Versions []FoundryVersionModel `tfsdk:"versions"`
}

type FoundryVersionModel struct {
	Name     types.String   `tfsdk:"name"`
	Channels []types.String `tfsdk:"channels"`
	EOLDate  types.String   `tfsdk:"eol_date"`
	Revoked  types.Bool     `tfsdk:"revoked"`
}

type NestVersionsDataSourceModel struct {
	Versions []NestVersionModel `tfsdk:"versions"`
}

type NestVersionModel struct {
	Name              types.String `tfsdk:"name"`
	KubernetesVersion types.String `tfsdk:"kubernetes_version"`
	EOLDate           types.String `tfsdk:"eol_date"`
	Revoked           types.Bool   `tfsdk:"revoked"`
}

func versionAttributes(kind string, extra map[string]schema.Attribute) map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{
		"name":     schema.StringAttribute{Computed: true, MarkdownDescription: fmt.Sprintf("Name of the %s version, as its `version` takes it", kind)},
		"eol_date": schema.StringAttribute{Computed: true, MarkdownDescription: "Date the version stops being supported"},
		"revoked":  schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the version was withdrawn. A revoked version cannot be updated to"},
	}
	for name, attribute := range extra {
		attributes[name] = attribute
	}
	return map[string]schema.Attribute{
		"versions": schema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: fmt.Sprintf("The %s versions this installation knows, newest first", kind),
			NestedObject:        schema.NestedAttributeObject{Attributes: attributes},
		},
	}
}

func (d *FoundryVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_foundry_versions"
}

func (d *NestVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest_versions"
}

func (d *FoundryVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Foundry versions a `meltcloud_foundry` can run. Needs the API Key of the admin Organization.",
		Attributes: versionAttributes("Foundry", map[string]schema.Attribute{
			"channels": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Release channels the version is on, for example `stable`"},
		}),
	}
}

func (d *NestVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Nest versions a `meltcloud_nest` can run. Needs the API Key of the admin Organization.",
		Attributes: versionAttributes("Nest", map[string]schema.Attribute{
			"kubernetes_version": schema.StringAttribute{Computed: true, MarkdownDescription: "Kubernetes version the Nest runs with it"},
		}),
	}
}

func configureClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *client.Client {
	if req.ProviderData == nil {
		return nil
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	return client
}

func (d *FoundryVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *NestVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *FoundryVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.FoundryVersions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Foundry versions, got error: %s", err))
		return
	}

	data := FoundryVersionsDataSourceModel{Versions: []FoundryVersionModel{}}
	for _, version := range result.FoundryVersions {
		channels := []types.String{}
		for _, channel := range version.Channels {
			channels = append(channels, types.StringValue(channel))
		}
		data.Versions = append(data.Versions, FoundryVersionModel{
			Name:     types.StringValue(version.Name),
			Channels: channels,
			EOLDate:  types.StringValue(version.EOLDate),
			Revoked:  types.BoolValue(version.Revoked),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NestVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.NestVersions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest versions, got error: %s", err))
		return
	}

	data := NestVersionsDataSourceModel{Versions: []NestVersionModel{}}
	for _, version := range result.NestVersions {
		data.Versions = append(data.Versions, NestVersionModel{
			Name:              types.StringValue(version.Name),
			KubernetesVersion: types.StringValue(version.KubernetesVersion),
			EOLDate:           types.StringValue(version.EOLDate),
			Revoked:           types.BoolValue(version.Revoked),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

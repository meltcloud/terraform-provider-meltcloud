package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NetworkProfileDataSource{}

func NewNetworkProfileDataSource() datasource.DataSource {
	return &NetworkProfileDataSource{}
}

// NetworkProfileDataSource defines the data source implementation.
type NetworkProfileDataSource struct {
	client *client.Client
}

// NetworkProfileDataSourceModel describes the data source data model.
type NetworkProfileDataSourceModel struct {
	ID     types.Int64           `tfsdk:"id"`
	Name   types.String          `tfsdk:"name"`
	Status types.String          `tfsdk:"status"`
	Links  []LinkDataSourceModel `tfsdk:"links"`
}

type LinkDataSourceModel struct {
	Name           types.String `tfsdk:"name"`
	Interfaces     types.List   `tfsdk:"interfaces"`
	VLANs          types.List   `tfsdk:"vlans"`
	HostNetworking types.Bool   `tfsdk:"host_networking"`
	LACP           types.Bool   `tfsdk:"lacp"`
	NativeVLAN     types.Bool   `tfsdk:"native_vlan"`
}

func (d *NetworkProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_profile"
}

func (d *NetworkProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: networkProfileDesc,

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: networkProfileResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: networkProfileResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the Network Profile",
				Computed:            true,
			},
			"links": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: linkResourceAttributes()["name"].GetMarkdownDescription(),
						},
						"interfaces": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: linkResourceAttributes()["interfaces"].GetMarkdownDescription(),
						},
						"vlans": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.Int64Type,
							MarkdownDescription: linkResourceAttributes()["vlans"].GetMarkdownDescription(),
						},
						"host_networking": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: linkResourceAttributes()["host_networking"].GetMarkdownDescription(),
						},
						"lacp": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: linkResourceAttributes()["lacp"].GetMarkdownDescription(),
						},
						"native_vlan": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: linkResourceAttributes()["native_vlan"].GetMarkdownDescription(),
						},
					},
				},
			},
		},
	}
}

func (d *NetworkProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.NetworkProfile().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network profile with ID %d , got error: %s", data.ID.ValueInt64(), err))
		return
	}

	for _, link := range result.NetworkProfile.Links {
		interfacesList, diags := types.ListValueFrom(ctx, types.StringType, link.Interfaces)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		vlansList, diags := types.ListValueFrom(ctx, types.Int64Type, link.VLANs)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		data.Links = append(data.Links, LinkDataSourceModel{
			Name:           types.StringValue(link.Name),
			Interfaces:     interfacesList,
			VLANs:          vlansList,
			HostNetworking: types.BoolValue(link.HostNetworking),
			LACP:           types.BoolValue(link.LACP),
			NativeVLAN:     types.BoolValue(link.NativeVLAN),
		})
	}

	data.ID = types.Int64Value(result.NetworkProfile.ID)
	data.Name = types.StringValue(result.NetworkProfile.Name)
	data.Status = types.StringValue(result.NetworkProfile.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
var _ datasource.DataSource = &MachinePoolDataSource{}

func NewMachinePoolDataSource() datasource.DataSource {
	return &MachinePoolDataSource{}
}

// MachinePoolDataSource defines the data source implementation.
type MachinePoolDataSource struct {
	client *client.Client
}

// MachinePoolDataSourceModel describes the data source data model.
type MachinePoolDataSourceModel struct {
	ID                    types.Int64                           `tfsdk:"id"`
	ClusterID             types.Int64                           `tfsdk:"cluster_id"`
	Name                  types.String                          `tfsdk:"name"`
	PrimaryDiskDevice     types.String                          `tfsdk:"primary_disk_device"`
	ReuseRootPartition    types.Bool                            `tfsdk:"reuse_existing_root_partition"`
	Version               types.String                          `tfsdk:"version"`
	PatchVersion          types.String                          `tfsdk:"patch_version"`
	Status                types.String                          `tfsdk:"status"`
	NetworkConfigurations []NetworkConfigurationDataSourceModel `tfsdk:"network_configurations"`
}

type NetworkConfigurationDataSourceModel struct {
	Type       types.String `tfsdk:"type"`
	Interfaces types.String `tfsdk:"interfaces"`
	VLANMode   types.String `tfsdk:"vlan_mode"`
	VLANs      types.String `tfsdk:"vlans"`
}

func (d *MachinePoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_pool"
}

func (d *MachinePoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: machinePoolDesc,

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: machinePoolResourceAttributes()["id"].GetMarkdownDescription(),
				Required:            true,
			},
			"cluster_id": schema.Int64Attribute{
				MarkdownDescription: machinePoolResourceAttributes()["cluster_id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: machinePoolResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"primary_disk_device": schema.StringAttribute{
				MarkdownDescription: machinePoolResourceAttributes()["primary_disk_device"].GetMarkdownDescription(),
				Computed:            true,
			},
			"reuse_existing_root_partition": schema.BoolAttribute{
				MarkdownDescription: machinePoolResourceAttributes()["reuse_existing_root_partition"].GetMarkdownDescription(),
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: machinePoolResourceAttributes()["version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"patch_version": schema.StringAttribute{
				MarkdownDescription: machinePoolResourceAttributes()["patch_version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the Machine Pool",
				Computed:            true,
			},
			"network_configurations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: networkConfigurationResourceAttributes()["type"].GetMarkdownDescription(),
						},

						"interfaces": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: networkConfigurationResourceAttributes()["interfaces"].GetMarkdownDescription(),
						},

						"vlan_mode": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: networkConfigurationResourceAttributes()["vlan_mode"].GetMarkdownDescription(),
						},

						"vlans": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: networkConfigurationResourceAttributes()["vlans"].GetMarkdownDescription(),
						},
					},
				},
			},
		},
	}
}

func (d *MachinePoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MachinePoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MachinePoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.MachinePool().Get(ctx, data.ClusterID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine pool with ID %d on cluster ID %d, got error: %s", data.ID.ValueInt64(), data.ClusterID.ValueInt64(), err))
		return
	}

	for _, networkConfiguration := range result.MachinePool.NetworkConfigurations {
		data.NetworkConfigurations = append(data.NetworkConfigurations, NetworkConfigurationDataSourceModel{
			Type:       types.StringValue(networkConfiguration.Type),
			Interfaces: types.StringValue(networkConfiguration.Interfaces),
			VLANMode:   types.StringValue(networkConfiguration.VLANMode),
			VLANs:      types.StringValue(networkConfiguration.VLANs),
		})
	}

	data.ID = types.Int64Value(result.MachinePool.ID)
	data.Name = types.StringValue(result.MachinePool.Name)
	data.PrimaryDiskDevice = types.StringValue(result.MachinePool.PrimaryDiskDevice)
	data.ReuseRootPartition = types.BoolValue(result.MachinePool.ReuseRootPartition)
	data.Version = types.StringValue(result.MachinePool.UserVersion)
	data.PatchVersion = types.StringValue(result.MachinePool.PatchVersion)
	data.Status = types.StringValue(result.MachinePool.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

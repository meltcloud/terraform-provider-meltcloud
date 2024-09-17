package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"strings"
	"terraform-provider-meltcloud/internal/client"
	"terraform-provider-meltcloud/internal/kubernetes"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClusterDataSource{}

func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

// ClusterDataSource defines the data source implementation.
type ClusterDataSource struct {
	client *client.Client
}

type ClusterDataSourceModel struct {
	ID                 types.Int64                `tfsdk:"id"`
	Name               types.String               `tfsdk:"name"`
	Version            types.String               `tfsdk:"version"`
	ControlPlaneStatus types.String               `tfsdk:"control_plane_status"`
	PatchVersion       types.String               `tfsdk:"patch_version"`
	PodCIDR            types.String               `tfsdk:"pod_cidr"`
	ServiceCIDR        types.String               `tfsdk:"service_cidr"`
	DNSServiceIP       types.String               `tfsdk:"dns_service_ip"`
	KubeConfigRaw      types.String               `tfsdk:"kubeconfig_raw"`
	KubeConfig         *KubeConfigDataSourceModel `tfsdk:"kubeconfig"`
}

type KubeConfigDataSourceModel struct {
	Host                 types.String `tfsdk:"host"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	ClientCertificate    types.String `tfsdk:"client_certificate"`
	ClientKey            types.String `tfsdk:"client_key"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
}

func (d *ClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *ClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: clusterDesc,

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: clusterResourceAttributes()["id"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: clusterResourceAttributes()["name"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"control_plane_status": schema.StringAttribute{
				MarkdownDescription: "Control Plane Status of the Cluster",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: clusterResourceAttributes()["version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"patch_version": schema.StringAttribute{
				MarkdownDescription: clusterResourceAttributes()["patch_version"].GetMarkdownDescription(),
				Computed:            true,
			},
			"pod_cidr": schema.StringAttribute{
				MarkdownDescription: clusterResourceAttributes()["pod_cidr"].GetMarkdownDescription(),
				Computed:            true,
			},
			"service_cidr": schema.StringAttribute{
				MarkdownDescription: clusterResourceAttributes()["service_cidr"].GetMarkdownDescription(),
				Computed:            true,
			},
			"dns_service_ip": schema.StringAttribute{
				MarkdownDescription: clusterResourceAttributes()["dns_service_ip"].GetMarkdownDescription(),
				Computed:            true,
			},
			"kubeconfig": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Computed:  true,
						Sensitive: true,
					},
					"username": schema.StringAttribute{
						Computed:  true,
						Sensitive: true,
					},
					"password": schema.StringAttribute{
						Computed:  true,
						Sensitive: true,
					},
					"client_certificate": schema.StringAttribute{
						Computed:  true,
						Sensitive: true,
					},
					"client_key": schema.StringAttribute{
						Computed:  true,
						Sensitive: true,
					},
					"cluster_ca_certificate": schema.StringAttribute{
						Computed:  true,
						Sensitive: true,
					},
				},
				Computed:  true,
				Sensitive: true,
			},
			"kubeconfig_raw": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (d *ClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cluster *client.Cluster
	if data.ID.ValueInt64() != 0 {
		result, err := d.client.Cluster().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster by ID %d, got error: %s", data.ID.ValueInt64(), err))
			return
		}
		cluster = result.Cluster
	} else {
		result, err := d.client.Cluster().List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read clusters, got error: %s", err))
			return
		}

		for _, c := range result.Clusters {
			if strings.EqualFold(data.Name.ValueString(), c.Name) {
				cluster = c
				break
			}
		}

		if cluster == nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not find cluster by name %s", data.Name.ValueString()))
			return
		}

		// need to lookup by ID since the List does not include the kubeconfig
		clusterResult, err2 := d.client.Cluster().Get(ctx, cluster.ID)
		if err2 != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster by ID %d, got error: %s", data.ID.ValueInt64(), err2))
			return
		}
		cluster = clusterResult.Cluster
	}

	data.ID = types.Int64Value(cluster.ID)
	data.Name = types.StringValue(cluster.Name)
	data.Version = types.StringValue(cluster.UserVersion)
	data.ControlPlaneStatus = types.StringValue(cluster.ControlPlaneStatus)
	data.PatchVersion = types.StringValue(cluster.PatchVersion)
	data.PodCIDR = types.StringValue(cluster.PodCIDR)
	data.ServiceCIDR = types.StringValue(cluster.ServiceCIDR)
	data.DNSServiceIP = types.StringValue(cluster.DNSServiceIP)
	data.KubeConfigRaw = types.StringValue(cluster.KubeConfig)

	kubeConfigDataModel, kErr := d.getKubeConfigResourceModel(cluster.KubeConfig)
	if kErr != nil {
		resp.Diagnostics.AddError("Client Error", kErr.Error())
		return
	}

	data.KubeConfig = kubeConfigDataModel
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *ClusterDataSource) getKubeConfigResourceModel(kubeconfig string) (*KubeConfigDataSourceModel, error) {
	kubeConfig, err := kubernetes.ParseKubeConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig error %+v", err)
	}

	return &KubeConfigDataSourceModel{
		Host:                 types.StringValue(kubeConfig.Clusters[0].Cluster.Server),
		Username:             types.StringValue(kubeConfig.Users[0].Name),
		Password:             types.StringValue(""),
		ClientCertificate:    types.StringValue(kubeConfig.Users[0].User.ClientCertificateData),
		ClientKey:            types.StringValue(kubeConfig.Users[0].User.ClientKeyData),
		ClusterCACertificate: types.StringValue(kubeConfig.Clusters[0].Cluster.ClusterAuthorityData),
	}, nil
}

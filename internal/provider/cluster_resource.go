package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-meltcloud/internal/client"
	"terraform-provider-meltcloud/internal/kubernetes"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	client *client.Client
}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Version       types.String `tfsdk:"version"`
	PatchVersion  types.String `tfsdk:"patch_version"`
	PodCIDR       types.String `tfsdk:"pod_cidr"`
	ServiceCIDR   types.String `tfsdk:"service_cidr"`
	DNSServiceIP  types.String `tfsdk:"dns_service_ip"`
	KubeConfigRaw types.String `tfsdk:"kubeconfig_raw"`
	KubeConfig    types.Object `tfsdk:"kubeconfig"`
}

type KubeConfigResourceModel struct {
	Host                 types.String `tfsdk:"host"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	ClientCertificate    types.String `tfsdk:"client_certificate"`
	ClientKey            types.String `tfsdk:"client_key"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

const clusterDesc string = "A [Cluster](https://meltcloud.io/docs/guides/clusters/create.html) in meltcloud consists of a **Kubernetes Control Plane** and associated objects like [Machine Pools](https://meltcloud.io/docs/guides/machine-pools/create.html) (which hold assigned [Machines](https://meltcloud.io/docs/guides/machine-pools/intro.html))."

func clusterResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Cluster in meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the cluster, not case-sensitive. Must be unique within the organization and consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com')",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes minor version of the cluster control plane",
			Required:            true,
		},
		"patch_version": schema.StringAttribute{
			MarkdownDescription: "Kubernetes patch version of the cluster control plane",
			Computed:            true,
		},
		"pod_cidr": schema.StringAttribute{
			MarkdownDescription: "CIDR for the Kubernetes Pods",
			Required:            true,
		},
		"service_cidr": schema.StringAttribute{
			MarkdownDescription: "CIDR for the Kubernetes Services",
			Required:            true,
		},
		"dns_service_ip": schema.StringAttribute{
			MarkdownDescription: "IP for the DNS service",
			Required:            true,
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
	}
}

func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: clusterDesc,

		Attributes: clusterResourceAttributes(),
	}
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clusterCreateInput := &client.ClusterCreateInput{
		Name:         data.Name.ValueString(),
		UserVersion:  data.Version.ValueString(),
		PodCIDR:      data.PodCIDR.ValueString(),
		ServiceCIDR:  data.ServiceCIDR.ValueString(),
		DNSServiceIP: data.DNSServiceIP.ValueString(),
	}

	clusterCreateResult, err := r.client.Cluster().Create(ctx, clusterCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", err))
		return
	}
	if clusterCreateResult.Operation == nil {
		resp.Diagnostics.AddError("Server Error", "Created cluster, but did not get operation")
		return
	}

	_, err = r.client.Operation().PollUntilDone(ctx, clusterCreateResult.Operation.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error during creation of cluster, got error: %s", err))
		return
	}

	clusterGetResult, err := r.client.Cluster().Get(ctx, clusterCreateResult.Cluster.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(clusterGetResult.Cluster.ID)
	data.PatchVersion = types.StringValue(clusterGetResult.Cluster.PatchVersion)
	data.KubeConfigRaw = types.StringValue(clusterGetResult.Cluster.KubeConfig)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	kubeConfigResourceModel, kErr := r.getKubeConfigResourceModel(clusterGetResult.Cluster.KubeConfig)
	if kErr != nil {
		resp.Diagnostics.AddError("Client Error", kErr.Error())
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("kubeconfig"), kubeConfigResourceModel)
	resp.Diagnostics.Append(diags...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Cluster().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.Cluster.Name)
	data.Version = types.StringValue(result.Cluster.UserVersion)
	data.PatchVersion = types.StringValue(result.Cluster.PatchVersion)
	data.KubeConfigRaw = types.StringValue(result.Cluster.KubeConfig)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	kubeConfigResourceModel, kErr := r.getKubeConfigResourceModel(result.Cluster.KubeConfig)
	if kErr != nil {
		resp.Diagnostics.AddError("Client Error", kErr.Error())
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("kubeconfig"), kubeConfigResourceModel)
	resp.Diagnostics.Append(diags...)
}

func (r *ClusterResource) getKubeConfigResourceModel(kubeconfig string) (*KubeConfigResourceModel, error) {
	kubeConfig, err := kubernetes.ParseKubeConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig error %+v", err)
	}

	return &KubeConfigResourceModel{
		Host:                 types.StringValue(kubeConfig.Clusters[0].Cluster.Server),
		Username:             types.StringValue(kubeConfig.Users[0].Name),
		Password:             types.StringValue(""),
		ClientCertificate:    types.StringValue(kubeConfig.Users[0].User.ClientCertificateData),
		ClientKey:            types.StringValue(kubeConfig.Users[0].User.ClientKeyData),
		ClusterCACertificate: types.StringValue(kubeConfig.Clusters[0].Cluster.ClusterAuthorityData),
	}, nil
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clusterUpdateInput := &client.ClusterUpdateInput{
		UserVersion: data.Version.ValueString(),
	}

	result, err := r.client.Cluster().Update(ctx, data.ID.ValueInt64(), clusterUpdateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", err))
		return
	}

	if result.Operation != nil {
		_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", err))
			return
		}

		_, err := r.client.Cluster().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
			return
		}
		data.PatchVersion = types.StringValue(result.Cluster.PatchVersion)
	} else {
		data.PatchVersion = types.StringValue(result.Cluster.PatchVersion)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Cluster().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete cluster, got error: %s", err))
		return
	}
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

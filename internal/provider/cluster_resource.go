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
	"regexp"
	"strconv"
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
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Version           types.String `tfsdk:"version"`
	PatchVersion      types.String `tfsdk:"patch_version"`
	PodCIDR           types.String `tfsdk:"pod_cidr"`
	ServiceCIDR       types.String `tfsdk:"service_cidr"`
	DNSServiceIP      types.String `tfsdk:"dns_service_ip"`
	AddonKubeProxy    types.Bool   `tfsdk:"addon_kube_proxy"`
	AddonCoreDNS      types.Bool   `tfsdk:"addon_core_dns"`
	KubeConfigRaw     types.String `tfsdk:"kubeconfig_raw"`
	KubeConfig        types.Object `tfsdk:"kubeconfig"`
	KubeConfigUserRaw types.String `tfsdk:"kubeconfig_user_raw"`
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

const clusterDesc string = "A [Cluster](https://docs.meltcloud.io/guides/clusters/create.html) in meltcloud consists of a **Kubernetes Control Plane** and associated objects like [Machine Pools](https://docs.meltcloud.io/guides/machine-pools/create.html) (which hold assigned [Machines](https://docs.meltcloud.io/guides/machine-pools/intro.html))."

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
		"addon_kube_proxy": schema.BoolAttribute{
			MarkdownDescription: "Enable kube-proxy Addon",
			Optional:            true,
			Computed:            true,
		},
		"addon_core_dns": schema.BoolAttribute{
			MarkdownDescription: "Enable CoreDNS Addon",
			Optional:            true,
			Computed:            true,
		},
		"kubeconfig": schema.SingleNestedAttribute{
			Description: "Kubeconfig values for the admin user",
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
			Description: "Kubeconfig file for the admin user",
			Computed:    true,
			Sensitive:   true,
		},
		"kubeconfig_user_raw": schema.StringAttribute{
			Description: "Kubeconfig file for the regular (OIDC) users",
			Computed:    true,
			Sensitive:   false,
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

	var addonKubeProxy *bool
	if !data.AddonKubeProxy.IsNull() && !data.AddonKubeProxy.IsUnknown() {
		addonKubeProxy = data.AddonKubeProxy.ValueBoolPointer()
	}

	var addonCoreDNS *bool
	if !data.AddonCoreDNS.IsNull() && !data.AddonCoreDNS.IsUnknown() {
		addonCoreDNS = data.AddonCoreDNS.ValueBoolPointer()
	}
	clusterCreateInput := &client.ClusterCreateInput{
		Name:           data.Name.ValueString(),
		UserVersion:    data.Version.ValueString(),
		PodCIDR:        data.PodCIDR.ValueString(),
		ServiceCIDR:    data.ServiceCIDR.ValueString(),
		DNSServiceIP:   data.DNSServiceIP.ValueString(),
		AddonKubeProxy: addonKubeProxy,
		AddonCoreDNS:   addonCoreDNS,
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
	data.Version = types.StringValue(clusterGetResult.Cluster.UserVersion)
	data.PatchVersion = types.StringValue(clusterGetResult.Cluster.PatchVersion)
	r.setValues(clusterGetResult.Cluster, &data)
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
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
		return
	}
	data.PodCIDR = types.StringValue(result.Cluster.PodCIDR)
	data.ServiceCIDR = types.StringValue(result.Cluster.ServiceCIDR)
	data.DNSServiceIP = types.StringValue(result.Cluster.DNSServiceIP)
	data.Version = types.StringValue(result.Cluster.UserVersion)
	data.PatchVersion = types.StringValue(result.Cluster.PatchVersion)
	r.setValues(result.Cluster, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	kubeConfigResourceModel, kErr := r.getKubeConfigResourceModel(result.Cluster.KubeConfig)
	if kErr != nil {
		resp.Diagnostics.AddError("Client Error", kErr.Error())
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("kubeconfig"), kubeConfigResourceModel)
	resp.Diagnostics.Append(diags...)
}

func (r *ClusterResource) setValues(result *client.Cluster, data *ClusterResourceModel) {
	data.Name = types.StringValue(result.Name)
	data.AddonKubeProxy = types.BoolValue(result.AddonKubeProxy)
	data.AddonCoreDNS = types.BoolValue(result.AddonCoreDNS)
	data.KubeConfigRaw = types.StringValue(result.KubeConfig)
	data.KubeConfigUserRaw = types.StringValue(result.KubeConfigUser)
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
	}
	r.setValues(result.Cluster, &data)
	data.PatchVersion = types.StringValue(result.Cluster.PatchVersion)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	kubeConfigResourceModel, kErr := r.getKubeConfigResourceModel(result.Cluster.KubeConfig)
	if kErr != nil {
		resp.Diagnostics.AddError("Client Error", kErr.Error())
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("kubeconfig"), kubeConfigResourceModel)
	resp.Diagnostics.Append(diags...)

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

var clusterImportIDPattern = regexp.MustCompile(`clusters/(\d+)`)

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := clusterImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", clusterImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

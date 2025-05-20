package client

import (
	"context"
	"fmt"
)

type ClusterRequest struct {
	client *Client
}

type ClusterResult struct {
	Cluster   *Cluster   `json:"cluster"`
	Operation *Operation `json:"operation,omitempty"`
}

type ClustersResult struct {
	Clusters []*Cluster `json:"clusters"`
}

type Cluster struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	ControlPlaneStatus string `json:"control_plane_status"`
	UserVersion        string `json:"user_version"`
	PatchVersion       string `json:"patch_version"`
	KubeConfig         string `json:"kubeconfig"`
	KubeConfigUser     string `json:"kubeconfig_user"`
	PodCIDR            string `json:"pod_cidr"`
	ServiceCIDR        string `json:"service_cidr"`
	DNSServiceIP       string `json:"dns_service_ip"`
	AddonKubeProxy     bool   `json:"addon_kube_proxy"`
	AddonCoreDNS       bool   `json:"addon_core_dns"`
}

type ClusterCreateInput struct {
	Name           string `json:"name"`
	UserVersion    string `json:"user_version"`
	PodCIDR        string `json:"pod_cidr"`
	ServiceCIDR    string `json:"service_cidr"`
	DNSServiceIP   string `json:"dns_service_ip"`
	AddonKubeProxy *bool  `json:"addon_kube_proxy,omitempty"`
	AddonCoreDNS   *bool  `json:"addon_core_dns,omitempty"`
}

type ClusterUpdateInput struct {
	UserVersion string `json:"user_version,omitempty"`
}

func (c *Client) Cluster() *ClusterRequest {
	return &ClusterRequest{
		client: c,
	}
}

func (mr *ClusterRequest) List(ctx context.Context) (*ClustersResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "clusters",
		Result: &ClustersResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clustersResult, ok := result.(*ClustersResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clustersResult, nil
}

func (mr *ClusterRequest) Get(ctx context.Context, id int64) (*ClusterResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "clusters", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &ClusterResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*ClusterResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *ClusterRequest) Create(ctx context.Context, input *ClusterCreateInput) (*ClusterResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "clusters",
		Result: &ClusterResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*ClusterResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *ClusterRequest) Update(ctx context.Context, id int64, input *ClusterUpdateInput) (*ClusterResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "clusters", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &ClusterResult{},
		Body:   input,
	}

	result, err := mr.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*ClusterResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *ClusterRequest) Delete(ctx context.Context, id int64) (*ClusterResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "clusters", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &ClusterResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*ClusterResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

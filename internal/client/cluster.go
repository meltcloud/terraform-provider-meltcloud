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

type Cluster struct {
	ID                 int64                     `json:"id"`
	Name               string                    `json:"name"`
	ControlPlaneStatus ClusterControlPlaneStatus `json:"control_plane_status"`
	UserVersion        string                    `json:"user_version"`
	PatchVersion       string                    `json:"patch_version"`
	KubeConfig         string                    `json:"kubeconfig"`
}

type ClusterCreateInput struct {
	Name         string `json:"name"`
	UserVersion  string `json:"user_version"`
	PodCIDR      string `json:"pod_cidr"`
	ServiceCIDR  string `json:"service_cidr"`
	DNSServiceIP string `json:"dns_service_ip"`
}

type ClusterUpdateInput struct {
	UserVersion string `json:"user_version,omitempty"`
}

type ClusterControlPlaneStatus string

const (
	ClusterStatusPending  ClusterControlPlaneStatus = "pending"
	ClusterStatusReady    ClusterControlPlaneStatus = "ready"
	ClusterStatusUpdating ClusterControlPlaneStatus = "updating"
)

func (c *Client) Cluster() *ClusterRequest {
	return &ClusterRequest{
		client: c,
	}
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

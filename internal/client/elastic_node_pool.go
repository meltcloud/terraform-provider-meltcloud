package client

import (
	"context"
	"fmt"
)

type ElasticNodePoolRequest struct {
	client *Client
}

type ElasticNodePoolResult struct {
	ElasticNodePool *ElasticNodePool `json:"elastic_node_pool"`
	Operation       *Operation       `json:"operation,omitempty"`
}

type ElasticNodePool struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	ClusterID      int64  `json:"cluster_id"`
	ElasticQuotaID int64  `json:"elastic_quota_id"`
	NodeCount      int64  `json:"node_count"`
	NodeVCPUs      int64  `json:"node_vcpus"`
	NodeMemoryMiB  int64  `json:"node_memory_mib"`
	NodeDiskGiB    int64  `json:"node_disk_gib"`
	Version        string `json:"version"`
	PatchVersion   string `json:"patch_version"`
}

type ElasticNodePoolCreateInput struct {
	Name           string `json:"name"`
	ElasticQuotaID int64  `json:"elastic_quota_id"`
	NodeCount      int64  `json:"node_count"`
	NodeVCPUs      int64  `json:"node_vcpus"`
	NodeMemoryMiB  int64  `json:"node_memory_mib"`
	NodeDiskGiB    int64  `json:"node_disk_gib"`
	Version        string `json:"version"`
}

type ElasticNodePoolUpdateInput struct {
	NodeCount     int64  `json:"node_count"`
	NodeVCPUs     int64  `json:"node_vcpus"`
	NodeMemoryMiB int64  `json:"node_memory_mib"`
	NodeDiskGiB   int64  `json:"node_disk_gib"`
	Version       string `json:"version"`
}

func (c *Client) ElasticNodePool() *ElasticNodePoolRequest {
	return &ElasticNodePoolRequest{
		client: c,
	}
}

func (er *ElasticNodePoolRequest) Get(ctx context.Context, clusterId int64, id int64) (*ElasticNodePoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "elastic_node_pools", id),
		Result: &ElasticNodePoolResult{},
	}

	result, err := er.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	nodePoolResult, ok := result.(*ElasticNodePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return nodePoolResult, nil
}

func (er *ElasticNodePoolRequest) Create(ctx context.Context, clusterId int64, input *ElasticNodePoolCreateInput) (*ElasticNodePoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s", "clusters", clusterId, "elastic_node_pools"),
		Result: &ElasticNodePoolResult{},
		Body:   input,
	}

	result, err := er.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	nodePoolResult, ok := result.(*ElasticNodePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return nodePoolResult, nil
}

func (er *ElasticNodePoolRequest) Update(ctx context.Context, clusterId int64, id int64, input *ElasticNodePoolUpdateInput) (*ElasticNodePoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "elastic_node_pools", id),
		Result: &ElasticNodePoolResult{},
		Body:   input,
	}

	result, err := er.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	nodePoolResult, ok := result.(*ElasticNodePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return nodePoolResult, nil
}

func (er *ElasticNodePoolRequest) Delete(ctx context.Context, clusterId int64, id int64) (*ElasticNodePoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "elastic_node_pools", id),
		Result: &ElasticNodePoolResult{},
	}

	result, err := er.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	nodePoolResult, ok := result.(*ElasticNodePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return nodePoolResult, nil
}

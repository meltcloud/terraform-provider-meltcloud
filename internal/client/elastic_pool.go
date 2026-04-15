package client

import (
	"context"
	"fmt"
)

type ElasticPoolRequest struct {
	client *Client
}

type ElasticPoolResult struct {
	ElasticPool *ElasticPool `json:"elastic_pool"`
	Operation   *Operation   `json:"operation,omitempty"`
}

type ElasticPool struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	ClusterID    int64  `json:"cluster_id"`
	ShareID      int64  `json:"share_id"`
	NodeCount    int64  `json:"node_count"`
	NodeCores    int64  `json:"node_cores"`
	NodeMemoryMB int64  `json:"node_memory_mb"`
	NodeDiskGB   int64  `json:"node_disk_gb"`
	Version      string `json:"version"`
	PatchVersion string `json:"patch_version"`
}

type ElasticPoolCreateInput struct {
	Name         string `json:"name"`
	ShareID      int64  `json:"share_id"`
	NodeCount    int64  `json:"node_count"`
	NodeCores    int64  `json:"node_cores"`
	NodeMemoryMB int64  `json:"node_memory_mb"`
	NodeDiskGB   int64  `json:"node_disk_gb"`
	Version      string `json:"version"`
}

type ElasticPoolUpdateInput struct {
	NodeCount    int64  `json:"node_count"`
	NodeCores    int64  `json:"node_cores"`
	NodeMemoryMB int64  `json:"node_memory_mb"`
	NodeDiskGB   int64  `json:"node_disk_gb"`
	Version      string `json:"version"`
}

func (c *Client) ElasticPool() *ElasticPoolRequest {
	return &ElasticPoolRequest{
		client: c,
	}
}

func (er *ElasticPoolRequest) Get(ctx context.Context, clusterId int64, id int64) (*ElasticPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "elastic_pools", id),
		Result: &ElasticPoolResult{},
	}

	result, err := er.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*ElasticPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

func (er *ElasticPoolRequest) Create(ctx context.Context, clusterId int64, input *ElasticPoolCreateInput) (*ElasticPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s", "clusters", clusterId, "elastic_pools"),
		Result: &ElasticPoolResult{},
		Body:   input,
	}

	result, err := er.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*ElasticPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

func (er *ElasticPoolRequest) Update(ctx context.Context, clusterId int64, id int64, input *ElasticPoolUpdateInput) (*ElasticPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "elastic_pools", id),
		Result: &ElasticPoolResult{},
		Body:   input,
	}

	result, err := er.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*ElasticPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

func (er *ElasticPoolRequest) Delete(ctx context.Context, clusterId int64, id int64) (*ElasticPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "elastic_pools", id),
		Result: &ElasticPoolResult{},
	}

	result, err := er.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*ElasticPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

package client

import (
	"context"
	"fmt"
)

type ElasticShareRequest struct {
	client *Client
}

type ElasticShareResult struct {
	ElasticShare *ElasticShare `json:"elastic_share"`
}

type ElasticShare struct {
	ID                        int64  `json:"id"`
	Name                      string `json:"name"`
	Cores                     int64  `json:"cores"`
	DiskGB                    int64  `json:"disk_gb"`
	MemoryMB                  int64  `json:"memory_mb"`
	CapacityID                int64  `json:"capacity_id"`
	ConsumingOrganizationUUID string `json:"consuming_organization_uuid"`
}

type ElasticShareCreateInput struct {
	Name                      string `json:"name"`
	Cores                     int64  `json:"cores"`
	DiskGB                    int64  `json:"disk_gb"`
	MemoryMB                  int64  `json:"memory_mb"`
	CapacityID                int64  `json:"capacity_id"`
	ConsumingOrganizationUUID string `json:"consuming_organization_uuid"`
}

type ElasticShareUpdateInput struct {
	Name     string `json:"name"`
	Cores    int64  `json:"cores"`
	DiskGB   int64  `json:"disk_gb"`
	MemoryMB int64  `json:"memory_mb"`
}

func (c *Client) ElasticShare() *ElasticShareRequest {
	return &ElasticShareRequest{
		client: c,
	}
}

func (er *ElasticShareRequest) Get(ctx context.Context, id int64) (*ElasticShareResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_shares", id),
		Result: &ElasticShareResult{},
	}

	result, err := er.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	shareResult, ok := result.(*ElasticShareResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return shareResult, nil
}

func (er *ElasticShareRequest) Create(ctx context.Context, input *ElasticShareCreateInput) (*ElasticShareResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "elastic_shares",
		Result: &ElasticShareResult{},
		Body:   input,
	}

	result, err := er.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	shareResult, ok := result.(*ElasticShareResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return shareResult, nil
}

func (er *ElasticShareRequest) Update(ctx context.Context, id int64, input *ElasticShareUpdateInput) (*ElasticShareResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_shares", id),
		Result: &ElasticShareResult{},
		Body:   input,
	}

	result, err := er.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	shareResult, ok := result.(*ElasticShareResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return shareResult, nil
}

func (er *ElasticShareRequest) Delete(ctx context.Context, id int64) (*ElasticShareResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_shares", id),
		Result: &ElasticShareResult{},
	}

	result, err := er.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	shareResult, ok := result.(*ElasticShareResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return shareResult, nil
}

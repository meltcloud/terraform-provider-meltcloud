package client

import (
	"context"
	"fmt"
)

type ElasticQuotaRequest struct {
	client *Client
}

type ElasticQuotaResult struct {
	ElasticQuota *ElasticQuota `json:"elastic_quota"`
}

type ElasticQuota struct {
	ID                        int64  `json:"id"`
	Name                      string `json:"name"`
	Cores                     int64  `json:"cores"`
	DiskGB                    int64  `json:"disk_gb"`
	MemoryMB                  int64  `json:"memory_mb"`
	ElasticFleetID            int64  `json:"elastic_fleet_id"`
	ConsumingOrganizationUUID string `json:"consuming_organization_uuid"`
}

type ElasticQuotaCreateInput struct {
	Name                      string `json:"name"`
	Cores                     int64  `json:"cores"`
	DiskGB                    int64  `json:"disk_gb"`
	MemoryMB                  int64  `json:"memory_mb"`
	ElasticFleetID            int64  `json:"elastic_fleet_id"`
	ConsumingOrganizationUUID string `json:"consuming_organization_uuid"`
}

type ElasticQuotaUpdateInput struct {
	Name     string `json:"name"`
	Cores    int64  `json:"cores"`
	DiskGB   int64  `json:"disk_gb"`
	MemoryMB int64  `json:"memory_mb"`
}

func (c *Client) ElasticQuota() *ElasticQuotaRequest {
	return &ElasticQuotaRequest{
		client: c,
	}
}

func (er *ElasticQuotaRequest) Get(ctx context.Context, id int64) (*ElasticQuotaResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_quotas", id),
		Result: &ElasticQuotaResult{},
	}

	result, err := er.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	quotaResult, ok := result.(*ElasticQuotaResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return quotaResult, nil
}

func (er *ElasticQuotaRequest) Create(ctx context.Context, input *ElasticQuotaCreateInput) (*ElasticQuotaResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "elastic_quotas",
		Result: &ElasticQuotaResult{},
		Body:   input,
	}

	result, err := er.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	quotaResult, ok := result.(*ElasticQuotaResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return quotaResult, nil
}

func (er *ElasticQuotaRequest) Update(ctx context.Context, id int64, input *ElasticQuotaUpdateInput) (*ElasticQuotaResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_quotas", id),
		Result: &ElasticQuotaResult{},
		Body:   input,
	}

	result, err := er.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	quotaResult, ok := result.(*ElasticQuotaResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return quotaResult, nil
}

func (er *ElasticQuotaRequest) Delete(ctx context.Context, id int64) (*ElasticQuotaResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_quotas", id),
		Result: &ElasticQuotaResult{},
	}

	result, err := er.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	quotaResult, ok := result.(*ElasticQuotaResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return quotaResult, nil
}

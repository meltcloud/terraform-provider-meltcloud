package client

import (
	"context"
	"fmt"
)

type ElasticCapacityRequest struct {
	client *Client
}

type ElasticCapacityResult struct {
	ElasticCapacity *ElasticCapacity `json:"elastic_capacity"`
	Operation       *Operation       `json:"operation,omitempty"`
}

type ElasticCapacity struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	ClusterID int64  `json:"cluster_id"`
}

type ElasticCapacityCreateInput struct {
	Name      string `json:"name"`
	ClusterID int64  `json:"cluster_id"`
}

func (c *Client) ElasticCapacity() *ElasticCapacityRequest {
	return &ElasticCapacityRequest{
		client: c,
	}
}

func (er *ElasticCapacityRequest) Get(ctx context.Context, id int64) (*ElasticCapacityResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_capacities", id),
		Result: &ElasticCapacityResult{},
	}

	result, err := er.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	capacityResult, ok := result.(*ElasticCapacityResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return capacityResult, nil
}

func (er *ElasticCapacityRequest) Create(ctx context.Context, input *ElasticCapacityCreateInput) (*ElasticCapacityResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "elastic_capacities",
		Result: &ElasticCapacityResult{},
		Body:   input,
	}

	result, err := er.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	capacityResult, ok := result.(*ElasticCapacityResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return capacityResult, nil
}

func (er *ElasticCapacityRequest) Delete(ctx context.Context, id int64) (*ElasticCapacityResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_capacities", id),
		Result: &ElasticCapacityResult{},
	}

	result, err := er.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	capacityResult, ok := result.(*ElasticCapacityResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return capacityResult, nil
}

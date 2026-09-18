package client

import (
	"context"
	"fmt"
	"net/url"
)

type ElasticFleetRequest struct {
	client *Client
}

type ElasticFleetResult struct {
	ElasticFleet *ElasticFleet `json:"elastic_fleet"`
	Operation    *Operation    `json:"operation,omitempty"`
}

type ElasticFleet struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	ClusterID int64  `json:"cluster_id"`
}

type ElasticFleetCreateInput struct {
	Name      string `json:"name"`
	ClusterID int64  `json:"cluster_id"`
}

func (c *Client) ElasticFleet() *ElasticFleetRequest {
	return &ElasticFleetRequest{
		client: c,
	}
}

func (er *ElasticFleetRequest) Get(ctx context.Context, id int64) (*ElasticFleetResult, *Error) {
	return er.get(ctx, &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_fleets", id),
		Result: &ElasticFleetResult{},
	})
}

func (er *ElasticFleetRequest) GetByName(ctx context.Context, name string) (*ElasticFleetResult, *Error) {
	return er.get(ctx, &ClientRequest{
		Path:        fmt.Sprintf("%s/%s", "elastic_fleets", url.PathEscape(name)),
		QueryParams: byName,
		Result:      &ElasticFleetResult{},
	})
}

func (er *ElasticFleetRequest) get(ctx context.Context, clientRequest *ClientRequest) (*ElasticFleetResult, *Error) {
	result, err := er.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	fleetResult, ok := result.(*ElasticFleetResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return fleetResult, nil
}

func (er *ElasticFleetRequest) Create(ctx context.Context, input *ElasticFleetCreateInput) (*ElasticFleetResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "elastic_fleets",
		Result: &ElasticFleetResult{},
		Body:   input,
	}

	result, err := er.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	fleetResult, ok := result.(*ElasticFleetResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return fleetResult, nil
}

func (er *ElasticFleetRequest) Delete(ctx context.Context, id int64) (*ElasticFleetResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "elastic_fleets", id),
		Result: &ElasticFleetResult{},
	}

	result, err := er.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	fleetResult, ok := result.(*ElasticFleetResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return fleetResult, nil
}

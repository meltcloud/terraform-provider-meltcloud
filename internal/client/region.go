package client

import (
	"context"
	"fmt"
	"net/url"
)

type RegionRequest struct {
	client *Client
}

type RegionResult struct {
	Region    *Region    `json:"region"`
	Operation *Operation `json:"operation,omitempty"`
}

type Region struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Zones []*Zone `json:"zones"`
}

type RegionCreateInput struct {
	Name string `json:"name"`
}

func (c *Client) Region() *RegionRequest {
	return &RegionRequest{client: c}
}

func (rr *RegionRequest) Get(ctx context.Context, id int64) (*RegionResult, *Error) {
	return rr.get(ctx, &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "regions", id),
		Result: &RegionResult{},
	})
}

func (rr *RegionRequest) GetByName(ctx context.Context, name string) (*RegionResult, *Error) {
	return rr.get(ctx, &ClientRequest{
		Path:        fmt.Sprintf("%s/%s", "regions", url.PathEscape(name)),
		QueryParams: byName,
		Result:      &RegionResult{},
	})
}

func (rr *RegionRequest) get(ctx context.Context, clientRequest *ClientRequest) (*RegionResult, *Error) {
	result, err := rr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	regionResult, ok := result.(*RegionResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return regionResult, nil
}

func (rr *RegionRequest) Create(ctx context.Context, input *RegionCreateInput) (*RegionResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "regions",
		Result: &RegionResult{},
		Body:   map[string]any{"region": input},
	}

	result, err := rr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	regionResult, ok := result.(*RegionResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return regionResult, nil
}

func (rr *RegionRequest) Delete(ctx context.Context, id int64) (*RegionResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "regions", id),
		Result: &RegionResult{},
	}

	result, err := rr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	regionResult, ok := result.(*RegionResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return regionResult, nil
}

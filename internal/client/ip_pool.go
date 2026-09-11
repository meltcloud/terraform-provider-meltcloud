package client

import (
	"context"
	"fmt"
	"net/url"
)

type IPPoolRequest struct {
	client *Client
}

type IPPoolResult struct {
	IPPool    *IPPool    `json:"ip_pool"`
	Operation *Operation `json:"operation,omitempty"`
}

type IPPool struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	CIDR        string        `json:"cidr"`
	Description *string       `json:"description"`
	Ranges      []IPPoolRange `json:"ranges"`
}

type IPPoolRange struct {
	Kind         string  `json:"kind"`
	StartAddress string  `json:"start_address"`
	EndAddress   string  `json:"end_address"`
	Description  *string `json:"description,omitempty"`
}

type IPPoolCreateInput struct {
	Name        string        `json:"name"`
	CIDR        string        `json:"cidr"`
	Description *string       `json:"description,omitempty"`
	Ranges      []IPPoolRange `json:"ranges,omitempty"`
}

type IPPoolUpdateInput struct {
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	Ranges      []IPPoolRange `json:"ranges"`
}

func (c *Client) IPPool() *IPPoolRequest {
	return &IPPoolRequest{client: c}
}

func (pr *IPPoolRequest) Get(ctx context.Context, id int64) (*IPPoolResult, *Error) {
	return pr.get(ctx, &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "ip_pools", id),
		Result: &IPPoolResult{},
	})
}

func (pr *IPPoolRequest) GetByName(ctx context.Context, name string) (*IPPoolResult, *Error) {
	return pr.get(ctx, &ClientRequest{
		Path:        fmt.Sprintf("%s/%s", "ip_pools", url.PathEscape(name)),
		QueryParams: byName,
		Result:      &IPPoolResult{},
	})
}

func (pr *IPPoolRequest) get(ctx context.Context, clientRequest *ClientRequest) (*IPPoolResult, *Error) {
	result, err := pr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*IPPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

func (pr *IPPoolRequest) Create(ctx context.Context, input *IPPoolCreateInput) (*IPPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "ip_pools",
		Result: &IPPoolResult{},
		Body:   map[string]any{"ip_pool": input},
	}

	result, err := pr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*IPPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

func (pr *IPPoolRequest) Update(ctx context.Context, id int64, input *IPPoolUpdateInput) (*IPPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "ip_pools", id),
		Result: &IPPoolResult{},
		Body:   map[string]any{"ip_pool": input},
	}

	result, err := pr.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*IPPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

func (pr *IPPoolRequest) Delete(ctx context.Context, id int64) (*IPPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "ip_pools", id),
		Result: &IPPoolResult{},
	}

	result, err := pr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	poolResult, ok := result.(*IPPoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return poolResult, nil
}

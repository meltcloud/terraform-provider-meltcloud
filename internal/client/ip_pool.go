package client

import (
	"context"
	"fmt"
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

// IPPoolRange is allocatable — where addresses are handed out from — or
// excluded, a hole inside one.
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

// IPPoolUpdateInput has no CIDR: every address handed out carries it.
type IPPoolUpdateInput struct {
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	Ranges      []IPPoolRange `json:"ranges"`
}

func (c *Client) IPPool() *IPPoolRequest {
	return &IPPoolRequest{client: c}
}

func (pr *IPPoolRequest) Get(ctx context.Context, id int64) (*IPPoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "ip_pools", id),
		Result: &IPPoolResult{},
	}

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

package client

import (
	"context"
	"fmt"
)

type NetworkRequest struct {
	client *Client
}

type NetworkResult struct {
	Network   *Network   `json:"network"`
	Operation *Operation `json:"operation,omitempty"`
}

type Network struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Subnets []*Subnet `json:"subnets,omitempty"`
}

type NetworkCreateInput struct {
	Name string `json:"name"`
}

func (c *Client) Network() *NetworkRequest {
	return &NetworkRequest{client: c}
}

func (nr *NetworkRequest) Get(ctx context.Context, id int64) (*NetworkResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "networks", id),
		Result: &NetworkResult{},
	}

	result, err := nr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	networkResult, ok := result.(*NetworkResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return networkResult, nil
}

func (nr *NetworkRequest) Create(ctx context.Context, input *NetworkCreateInput) (*NetworkResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "networks",
		Result: &NetworkResult{},
		Body:   map[string]any{"network": input},
	}

	result, err := nr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	networkResult, ok := result.(*NetworkResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return networkResult, nil
}

func (nr *NetworkRequest) Delete(ctx context.Context, id int64) (*NetworkResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d", "networks", id),
		Result: &NetworkResult{},
	}

	result, err := nr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	networkResult, ok := result.(*NetworkResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return networkResult, nil
}

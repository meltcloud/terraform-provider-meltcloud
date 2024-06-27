package client

import (
	"context"
	"fmt"
	"time"
)

type IPXEBootURLRequest struct {
	client *Client
}

type IPXEBootURLResult struct {
	IPXEBootURL *IPXEBootURL `json:"ipxe_boot_url"`
	Operation   *Operation   `json:"operation,omitempty"`
}

type IPXEBootURL struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
	BootURL   string    `json:"url"`
}

type IPXEBootURLCreateInput struct {
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Client) IPXEBootURL() *IPXEBootURLRequest {
	return &IPXEBootURLRequest{
		client: c,
	}
}

func (mr *IPXEBootURLRequest) Get(ctx context.Context, id int64) (*IPXEBootURLResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_boot_urls", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEBootURLResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEBootURLRequest) Create(ctx context.Context, input *IPXEBootURLCreateInput) (*IPXEBootURLResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "ipxe_boot_urls",
		Result: &IPXEBootURLResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEBootURLRequest) Delete(ctx context.Context, id int64) (*IPXEBootURLResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_boot_urls", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEBootURLResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

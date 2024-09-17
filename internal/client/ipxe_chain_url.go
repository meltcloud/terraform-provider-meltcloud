package client

import (
	"context"
	"fmt"
	"time"
)

type IPXEChainURLRequest struct {
	client *Client
}

type IPXEChainURLResult struct {
	IPXEChainURL *IPXEChainURL `json:"ipxe_chain_url"`
	Operation    *Operation    `json:"operation,omitempty"`
}

type IPXEChainURLsResult struct {
	IPXEChainURLs []*IPXEChainURL `json:"ipxe_chain_urls"`
}

type IPXEChainURL struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
	URL       string    `json:"url"`
	Script    string    `json:"script"`
}

type IPXEChainURLCreateInput struct {
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Client) IPXEChainURL() *IPXEChainURLRequest {
	return &IPXEChainURLRequest{
		client: c,
	}
}

func (mr *IPXEChainURLRequest) List(ctx context.Context) (*IPXEChainURLsResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "ipxe_chain_urls",
		Result: &IPXEChainURLsResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	iPXEBootArtifactsResult, ok := result.(*IPXEChainURLsResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return iPXEBootArtifactsResult, nil
}

func (mr *IPXEChainURLRequest) Get(ctx context.Context, id int64) (*IPXEChainURLResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_chain_urls", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEChainURLResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEChainURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEChainURLRequest) Create(ctx context.Context, input *IPXEChainURLCreateInput) (*IPXEChainURLResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "ipxe_chain_urls",
		Result: &IPXEChainURLResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEChainURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEChainURLRequest) Delete(ctx context.Context, id int64) (*IPXEChainURLResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_chain_urls", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEChainURLResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEChainURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

package client

import (
	"context"
	"fmt"
	"time"
)

type IPXEBootISORequest struct {
	client *Client
}

type IPXEBootISOResult struct {
	IPXEBootISO *IPXEBootISO `json:"ipxe_boot_iso"`
	Operation   *Operation   `json:"operation,omitempty"`
}

type IPXEBootISO struct {
	ID          int64             `json:"id"`
	ExpiresAt   time.Time         `json:"expires_at"`
	Status      IPXEBootISOStatus `json:"status"`
	DownloadURL string            `json:"download_url"`
}

type IPXEBootISOCreateInput struct {
	ExpiresAt time.Time `json:"expires_at"`
}

type IPXEBootISOStatus string

const (
	IPXEBootISOStatusPending IPXEBootISOStatus = "pending"
	IPXEBootISOStatusReady   IPXEBootISOStatus = "ready"
)

func (c *Client) IPXEBootISO() *IPXEBootISORequest {
	return &IPXEBootISORequest{
		client: c,
	}
}

func (mr *IPXEBootISORequest) Get(ctx context.Context, id int64) (*IPXEBootISOResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_boot_isos", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEBootISOResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootISOResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEBootISORequest) Create(ctx context.Context, input *IPXEBootISOCreateInput) (*IPXEBootISOResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "ipxe_boot_isos",
		Result: &IPXEBootISOResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootISOResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEBootISORequest) Delete(ctx context.Context, id int64) (*IPXEBootISOResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_boot_isos", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEBootISOResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootISOResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

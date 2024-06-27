package client

import (
	"context"
	"fmt"
	"time"
)

type IPXEBootArtifactRequest struct {
	client *Client
}

type IPXEBootArtifactResult struct {
	IPXEBootArtifact *IPXEBootArtifact `json:"ipxe_boot_artifact"`
	Operation        *Operation        `json:"operation,omitempty"`
}

type IPXEBootArtifact struct {
	ID             int64                  `json:"id"`
	Name           string                 `json:"name"`
	ExpiresAt      time.Time              `json:"expires_at"`
	Status         IPXEBootArtifactStatus `json:"status"`
	DownloadURLISO string                 `json:"download_url_iso"`
	DownloadURLPXE string                 `json:"download_url_pxe"`
	DownloadURLEFI string                 `json:"download_url_efi"`
}

type IPXEBootArtifactCreateInput struct {
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
}

type IPXEBootArtifactStatus string

const (
	IPXEBootArtifactStatusPending IPXEBootArtifactStatus = "pending"
	IPXEBootArtifactStatusReady   IPXEBootArtifactStatus = "ready"
)

func (c *Client) IPXEBootArtifact() *IPXEBootArtifactRequest {
	return &IPXEBootArtifactRequest{
		client: c,
	}
}

func (mr *IPXEBootArtifactRequest) Get(ctx context.Context, id int64) (*IPXEBootArtifactResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_boot_artifacts", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEBootArtifactResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootArtifactResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEBootArtifactRequest) Create(ctx context.Context, input *IPXEBootArtifactCreateInput) (*IPXEBootArtifactResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "ipxe_boot_artifacts",
		Result: &IPXEBootArtifactResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootArtifactResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *IPXEBootArtifactRequest) Delete(ctx context.Context, id int64) (*IPXEBootArtifactResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "ipxe_boot_artifacts", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &IPXEBootArtifactResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*IPXEBootArtifactResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

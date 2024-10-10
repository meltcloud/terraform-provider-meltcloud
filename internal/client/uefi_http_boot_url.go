package client

import (
	"context"
	"fmt"
	"time"
)

type UEFIHTTPBootURLRequest struct {
	client *Client
}

type UEFIHTTPBootURLResult struct {
	UEFIHTTPBootURL *UEFIHTTPBootURL `json:"uefi_http_boot_url"`
}

type UEFIHTTPBootURLsResult struct {
	UEFIHTTPBootURLs []*UEFIHTTPBootURL `json:"uefi_http_boot_urls"`
}

type UEFIHTTPBootURL struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	ExpiresAt     time.Time `json:"expires_at"`
	Protocols     string    `json:"protocols"`
	HTTPURLAMD64  string    `json:"http_url_amd64"`
	HTTPSURLAMD64 string    `json:"https_url_amd64"`
	HTTPURLARM64  string    `json:"http_url_arm64"`
	HTTPSURLARM64 string    `json:"https_url_arm64"`
}

type UEFIHTTPBootURLCreateInput struct {
	Name      string    `json:"name"`
	Protocols string    `json:"protocols"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Client) UEFIHTTPBootURL() *UEFIHTTPBootURLRequest {
	return &UEFIHTTPBootURLRequest{
		client: c,
	}
}

func (mr *UEFIHTTPBootURLRequest) List(ctx context.Context, iPXEBootArtifactID int64) (*UEFIHTTPBootURLsResult, *Error) {
	subPath := fmt.Sprintf("%s/%d/%s", "ipxe_boot_artifacts", iPXEBootArtifactID, "uefi_http_boot_urls")
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &UEFIHTTPBootURLsResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	iPXEBootArtifactsResult, ok := result.(*UEFIHTTPBootURLsResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return iPXEBootArtifactsResult, nil
}

func (mr *UEFIHTTPBootURLRequest) Get(ctx context.Context, iPXEBootArtifactID int64, id int64) (*UEFIHTTPBootURLResult, *Error) {
	subPath := fmt.Sprintf("%s/%d/%s/%d", "ipxe_boot_artifacts", iPXEBootArtifactID, "uefi_http_boot_urls", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &UEFIHTTPBootURLResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*UEFIHTTPBootURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *UEFIHTTPBootURLRequest) Create(ctx context.Context, iPXEBootArtifactID int64, input *UEFIHTTPBootURLCreateInput) (*UEFIHTTPBootURLResult, *Error) {
	subPath := fmt.Sprintf("%s/%d/%s/", "ipxe_boot_artifacts", iPXEBootArtifactID, "uefi_http_boot_urls")
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &UEFIHTTPBootURLResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*UEFIHTTPBootURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *UEFIHTTPBootURLRequest) Delete(ctx context.Context, iPXEBootArtifactID int64, id int64) (*UEFIHTTPBootURLResult, *Error) {
	subPath := fmt.Sprintf("%s/%d/%s/%d", "ipxe_boot_artifacts", iPXEBootArtifactID, "uefi_http_boot_urls", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &UEFIHTTPBootURLResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*UEFIHTTPBootURLResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

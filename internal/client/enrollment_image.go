package client

import (
	"context"
	"fmt"
	"time"
)

type EnrollmentImageRequest struct {
	client *Client
}

type EnrollmentImageResult struct {
	EnrollmentImage *EnrollmentImage `json:"enrollment_image"`
	Operation       *Operation       `json:"operation,omitempty"`
}

type EnrollmentImagesResult struct {
	EnrollmentImages []*EnrollmentImage `json:"enrollment_images"`
}

type EnrollmentImage struct {
	ID                        int64      `json:"id"`
	Name                      string     `json:"name"`
	ExpiresAt                 time.Time  `json:"expires_at"`
	Status                    string     `json:"status"`
	InstallDiskDevice         string     `json:"install_disk_device"`
	InstallDiskForceOverwrite bool       `json:"install_disk_force_overwrite"`
	VLAN                      *int64     `json:"vlan"`
	EnableHTTP                bool       `json:"enable_http"`
	HTTPURLISOAMD64           string     `json:"http_url_iso_amd64"`
	HTTPURLISOARM64           string     `json:"http_url_iso_arm64"`
	HTTPSURLISOAMD64          string     `json:"https_url_iso_arm64"`
	HTTPSURLISOARM64          string     `json:"https_url_iso_amd64"`
	IPXEScriptHTTPAMD64       string     `json:"ipxe_script_http_amd64"`
	IPXEScriptHTTPARM64       string     `json:"ipxe_script_http_arm64"`
	IPXEScriptHTTPSAMD64      string     `json:"ipxe_script_https_amd64"`
	IPXEScriptHTTPSARM64      string     `json:"ipxe_script_https_arm64"`
	LastUsedAt                *time.Time `json:"last_used_at"`
}

type EnrollmentImageCreateInput struct {
	Name                      string    `json:"name"`
	ExpiresAt                 time.Time `json:"expires_at"`
	InstallDiskDevice         string    `json:"install_disk_device"`
	InstallDiskForceOverwrite *bool     `json:"install_disk_force_overwrite"`
	VLAN                      *int64    `json:"vlan"`
	EnableHTTP                *bool     `json:"enable_http"`
}

func (c *Client) EnrollmentImage() *EnrollmentImageRequest {
	return &EnrollmentImageRequest{
		client: c,
	}
}

func (mr *EnrollmentImageRequest) List(ctx context.Context) (*EnrollmentImagesResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "enrollment_images",
		Result: &EnrollmentImagesResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	iPXEBootArtifactsResult, ok := result.(*EnrollmentImagesResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return iPXEBootArtifactsResult, nil
}

func (mr *EnrollmentImageRequest) Get(ctx context.Context, id int64) (*EnrollmentImageResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "enrollment_images", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &EnrollmentImageResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*EnrollmentImageResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *EnrollmentImageRequest) Create(ctx context.Context, input *EnrollmentImageCreateInput) (*EnrollmentImageResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "enrollment_images",
		Result: &EnrollmentImageResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*EnrollmentImageResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

func (mr *EnrollmentImageRequest) Delete(ctx context.Context, id int64) (*EnrollmentImageResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "enrollment_images", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &EnrollmentImageResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	clusterResult, ok := result.(*EnrollmentImageResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return clusterResult, nil
}

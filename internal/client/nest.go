package client

import (
	"context"
	"fmt"
	"net/url"
)

type NestRequest struct {
	client *Client
}

type NestResult struct {
	Nest      *Nest      `json:"nest"`
	Operation *Operation `json:"operation,omitempty"`
}

type Nest struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Revision int64  `json:"revision"`
	Version  string `json:"version"`
}

type NestInstanceResult struct {
	NestInstance *NestInstance `json:"nest_instance"`
}

type NestInstance struct {
	ID                      int64              `json:"id"`
	NestID                  int64              `json:"nest_id"`
	Name                    string             `json:"name"`
	Status                  string             `json:"status"`
	ZoneID                  int64              `json:"zone_id"`
	IP                      string             `json:"ip"`
	Architecture            string             `json:"architecture"`
	Version                 string             `json:"version"`
	InstallDiskDevice       string             `json:"install_disk_device"`
	InstallDiskMirror       bool               `json:"install_disk_mirror"`
	InstallDiskMirrorDevice string             `json:"install_disk_mirror_device"`
	EtcdDisk                bool               `json:"etcd_disk"`
	EtcdDiskDevice          string             `json:"etcd_disk_device"`
	EtcdDiskMirror          bool               `json:"etcd_disk_mirror"`
	EtcdDiskMirrorDevice    string             `json:"etcd_disk_mirror_device"`
	Image                   *NestInstanceImage `json:"image"`
}

// The install image. Its URLs are set once it is built, and carry its download token.
type NestInstanceImage struct {
	Status      string `json:"status"`
	URLISOAMD64 string `json:"url_iso_amd64"`
	URLISOARM64 string `json:"url_iso_arm64"`
}

type NestInstanceCreateInput struct {
	Name                    string  `json:"name"`
	IP                      string  `json:"ip"`
	ZoneID                  int64   `json:"zone_id"`
	InstallDiskDevice       *string `json:"install_disk_device,omitempty"`
	InstallDiskMirror       *bool   `json:"install_disk_mirror,omitempty"`
	InstallDiskMirrorDevice *string `json:"install_disk_mirror_device,omitempty"`
	EtcdDisk                *bool   `json:"etcd_disk,omitempty"`
	EtcdDiskDevice          *string `json:"etcd_disk_device,omitempty"`
	EtcdDiskMirror          *bool   `json:"etcd_disk_mirror,omitempty"`
	EtcdDiskMirrorDevice    *string `json:"etcd_disk_mirror_device,omitempty"`
}

func (c *Client) Nest() *NestRequest {
	return &NestRequest{client: c}
}

func (nr *NestRequest) Get(ctx context.Context) (*NestResult, *Error) {
	return send[NestResult](ctx, nr.client.Get, &ClientRequest{Path: "nest", Result: &NestResult{}})
}

// Rolls the Nest out to the version. The operation is nil if it runs that version already.
func (nr *NestRequest) Update(ctx context.Context, version string) (*NestResult, *Error) {
	return send[NestResult](ctx, nr.client.Put, &ClientRequest{
		Path:   "nest",
		Result: &NestResult{},
		Body:   map[string]any{"version": version},
	})
}

func (nr *NestRequest) GetInstance(ctx context.Context, id int64) (*NestInstanceResult, *Error) {
	return send[NestInstanceResult](ctx, nr.client.Get, &ClientRequest{
		Path:   fmt.Sprintf("nest_instances/%d", id),
		Result: &NestInstanceResult{},
	})
}

func (nr *NestRequest) GetInstanceByName(ctx context.Context, name string) (*NestInstanceResult, *Error) {
	return send[NestInstanceResult](ctx, nr.client.Get, &ClientRequest{
		Path:        fmt.Sprintf("nest_instances/%s", url.PathEscape(name)),
		QueryParams: byName,
		Result:      &NestInstanceResult{},
	})
}

func (nr *NestRequest) CreateInstance(ctx context.Context, input *NestInstanceCreateInput) (*NestInstanceResult, *Error) {
	return send[NestInstanceResult](ctx, nr.client.Post, &ClientRequest{
		Path:   "nest_instances",
		Result: &NestInstanceResult{},
		Body:   input,
	})
}

func send[T any](ctx context.Context, do func(context.Context, *ClientRequest) (any, *Error), clientRequest *ClientRequest) (*T, *Error) {
	result, err := do(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	typed, ok := result.(*T)
	if !ok {
		return nil, &ErrorTypeAssert
	}
	return typed, nil
}

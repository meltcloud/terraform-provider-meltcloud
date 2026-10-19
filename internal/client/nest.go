package client

import (
	"context"
	"fmt"
	"net/url"
	"time"
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
	ID                      int64       `json:"id"`
	NestID                  int64       `json:"nest_id"`
	Name                    string      `json:"name"`
	Status                  string      `json:"status"`
	ZoneID                  int64       `json:"zone_id"`
	NetworkProfileID        int64       `json:"network_profile_id"`
	IPAddresses             []IPAddress `json:"ip_addresses"`
	Architecture            string      `json:"architecture"`
	Version                 string      `json:"version"`
	InstallDiskDevice       string      `json:"install_disk_device"`
	InstallDiskMirror       bool        `json:"install_disk_mirror"`
	InstallDiskMirrorDevice string      `json:"install_disk_mirror_device"`
	EtcdDisk                bool        `json:"etcd_disk"`
	EtcdDiskDevice          string      `json:"etcd_disk_device"`
	EtcdDiskMirror          bool        `json:"etcd_disk_mirror"`
	EtcdDiskMirrorDevice    string      `json:"etcd_disk_mirror_device"`
	BGPLocalASN             *int64      `json:"bgp_local_asn"`
	BGPPeers                []BGPPeer   `json:"bgp_peers"`
}

// The API never returns the password.
type BGPPeer struct {
	IP                  string  `json:"bgp_peer_ip"`
	ASN                 int64   `json:"bgp_peer_asn"`
	Password            *string `json:"bgp_peer_password,omitempty"`
	Multihop            bool    `json:"bgp_multihop"`
	BFD                 bool    `json:"bgp_bfd"`
	BFDReceiveInterval  *int64  `json:"bgp_bfd_receive_interval,omitempty"`
	BFDTransmitInterval *int64  `json:"bgp_bfd_transmit_interval,omitempty"`
	BFDDetectMultiplier *int64  `json:"bgp_bfd_detect_multiplier,omitempty"`
}

type NestInstanceCreateInput struct {
	Name                    string      `json:"name"`
	ZoneID                  int64       `json:"zone_id"`
	NetworkProfileID        int64       `json:"network_profile_id"`
	IPAddresses             []IPAddress `json:"ip_addresses"`
	InstallDiskDevice       *string     `json:"install_disk_device,omitempty"`
	InstallDiskMirror       *bool       `json:"install_disk_mirror,omitempty"`
	InstallDiskMirrorDevice *string     `json:"install_disk_mirror_device,omitempty"`
	EtcdDisk                *bool       `json:"etcd_disk,omitempty"`
	EtcdDiskDevice          *string     `json:"etcd_disk_device,omitempty"`
	EtcdDiskMirror          *bool       `json:"etcd_disk_mirror,omitempty"`
	EtcdDiskMirrorDevice    *string     `json:"etcd_disk_mirror_device,omitempty"`
	BGPLocalASN             *int64      `json:"bgp_local_asn,omitempty"`
	BGPPeers                []BGPPeer   `json:"bgp_peers,omitempty"`
}

type NestInstanceImageResult struct {
	NestInstanceImage *NestInstanceImage `json:"nest_instance_image"`
	Operation         *Operation         `json:"operation,omitempty"`
}

// The URLs are set once the image is built, and carry its download token.
type NestInstanceImage struct {
	ID               int64      `json:"id"`
	NestInstanceID   int64      `json:"nest_instance_id"`
	Name             string     `json:"name"`
	Status           string     `json:"status"`
	ExpiresAt        time.Time  `json:"expires_at"`
	LastUsedAt       *time.Time `json:"last_used_at"`
	HTTPSURLISOAMD64 *string    `json:"https_url_iso_amd64"`
	HTTPSURLISOARM64 *string    `json:"https_url_iso_arm64"`
}

type NestInstanceImageCreateInput struct {
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
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

// Only an instance that never joined the Nest or was reset can be deleted.
func (nr *NestRequest) DeleteInstance(ctx context.Context, id int64) (*NestInstanceResult, *Error) {
	return send[NestInstanceResult](ctx, nr.client.Delete, &ClientRequest{
		Path:   fmt.Sprintf("nest_instances/%d", id),
		Result: &NestInstanceResult{},
	})
}

func (nr *NestRequest) GetInstanceImage(ctx context.Context, instanceID, id int64) (*NestInstanceImageResult, *Error) {
	return send[NestInstanceImageResult](ctx, nr.client.Get, &ClientRequest{
		Path:   fmt.Sprintf("nest_instances/%d/images/%d", instanceID, id),
		Result: &NestInstanceImageResult{},
	})
}

func (nr *NestRequest) GetInstanceImageByName(ctx context.Context, instanceID int64, name string) (*NestInstanceImageResult, *Error) {
	return send[NestInstanceImageResult](ctx, nr.client.Get, &ClientRequest{
		Path:        fmt.Sprintf("nest_instances/%d/images/%s", instanceID, url.PathEscape(name)),
		QueryParams: byName,
		Result:      &NestInstanceImageResult{},
	})
}

// Starts building the image; the operation is the build.
func (nr *NestRequest) CreateInstanceImage(ctx context.Context, instanceID int64, input *NestInstanceImageCreateInput) (*NestInstanceImageResult, *Error) {
	return send[NestInstanceImageResult](ctx, nr.client.Post, &ClientRequest{
		Path:   fmt.Sprintf("nest_instances/%d/images", instanceID),
		Result: &NestInstanceImageResult{},
		Body:   input,
	})
}

func (nr *NestRequest) DeleteInstanceImage(ctx context.Context, instanceID, id int64) (*NestInstanceImageResult, *Error) {
	return send[NestInstanceImageResult](ctx, nr.client.Delete, &ClientRequest{
		Path:   fmt.Sprintf("nest_instances/%d/images/%d", instanceID, id),
		Result: &NestInstanceImageResult{},
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

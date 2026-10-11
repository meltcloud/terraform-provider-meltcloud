package client

import (
	"context"
	"fmt"
	"net/url"
)

type ZoneRequest struct {
	client *Client
}

type ZoneResult struct {
	Zone      *Zone      `json:"zone"`
	Operation *Operation `json:"operation,omitempty"`
}

type Zone struct {
	ID       int64  `json:"id"`
	RegionID int64  `json:"region_id"`
	Name     string `json:"name"`
}

type ZoneCreateInput struct {
	Name string `json:"name"`
}

func (c *Client) Zone() *ZoneRequest {
	return &ZoneRequest{client: c}
}

func zonePath(regionID int64, parts ...any) string {
	path := fmt.Sprintf("%s/%d/%s", "regions", regionID, "zones")
	for _, part := range parts {
		path = fmt.Sprintf("%s/%v", path, part)
	}
	return path
}

func (zr *ZoneRequest) Get(ctx context.Context, regionID int64, id int64) (*ZoneResult, *Error) {
	return zr.do(ctx, zr.client.Get, &ClientRequest{Path: zonePath(regionID, id), Result: &ZoneResult{}})
}

func (zr *ZoneRequest) GetByName(ctx context.Context, regionID int64, name string) (*ZoneResult, *Error) {
	return zr.do(ctx, zr.client.Get, &ClientRequest{
		Path:        zonePath(regionID, url.PathEscape(name)),
		QueryParams: byName,
		Result:      &ZoneResult{},
	})
}

func (zr *ZoneRequest) Create(ctx context.Context, regionID int64, input *ZoneCreateInput) (*ZoneResult, *Error) {
	return zr.do(ctx, zr.client.Post, &ClientRequest{
		Path:   zonePath(regionID),
		Result: &ZoneResult{},
		Body:   map[string]any{"zone": input},
	})
}

func (zr *ZoneRequest) Delete(ctx context.Context, regionID int64, id int64) (*ZoneResult, *Error) {
	return zr.do(ctx, zr.client.Delete, &ClientRequest{Path: zonePath(regionID, id), Result: &ZoneResult{}})
}

func (zr *ZoneRequest) do(ctx context.Context, send func(context.Context, *ClientRequest) (any, *Error), clientRequest *ClientRequest) (*ZoneResult, *Error) {
	result, err := send(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	zoneResult, ok := result.(*ZoneResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return zoneResult, nil
}

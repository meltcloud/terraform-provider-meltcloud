package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const apiPath string = "/api/v1/"

const (
	requestTimeout       = 60 * time.Second
	rateLimitRetries     = 5
	maxRateLimitWait     = time.Minute
	defaultRateLimitWait = 5 * time.Second
)

type Client struct {
	Debug        bool
	HttpClient   *resty.Client
	Endpoint     string
	Organization string
}

type ClientRequest struct {
	Path        string
	QueryParams map[string]string
	Result      interface{}
	Body        interface{}
}

type Metadata struct {
	CurrentPage int `json:"current_page,omitempty"`
	NextPage    int `json:"next_page,omitempty"`
	PrevPage    int `json:"prev_page,omitempty"`
	TotalPages  int `json:"total_pages,omitempty"`
	TotalCount  int `json:"total_count,omitempty"`
}

type TLSConfig struct {
	CACertPEM     string
	SkipTLSVerify bool
}

func New(endpoint string, organization string, apiKey string, tlsConfig *TLSConfig) *Client {
	baseURL := fmt.Sprintf("%s%s/orgs/%s/", endpoint, apiPath, url.PathEscape(organization))

	// Don't follow redirects, so the API key is only sent to the endpoint.
	restyClient := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "meltcloud-go-client v1").
		SetHeader("X-Meltcloud-API-Key", apiKey).
		SetRedirectPolicy(resty.NoRedirectPolicy()).
		SetTimeout(requestTimeout).
		SetRetryCount(rateLimitRetries).
		SetRetryMaxWaitTime(maxRateLimitWait).
		AddRetryCondition(func(r *resty.Response, _ error) bool {
			return r != nil && r.StatusCode() == http.StatusTooManyRequests
		}).
		SetRetryAfter(retryAfter)

	if tlsConfig != nil {
		tc := &tls.Config{}
		if tlsConfig.SkipTLSVerify {
			tc.InsecureSkipVerify = true
		}
		if tlsConfig.CACertPEM != "" {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM([]byte(tlsConfig.CACertPEM))
			tc.RootCAs = pool
		}
		restyClient.SetTLSClientConfig(tc)
	}

	return &Client{
		HttpClient:   restyClient,
		Endpoint:     endpoint,
		Organization: organization,
	}
}

// retryAfter returns the wait from the Retry-After header, up to maxRateLimitWait.
func retryAfter(_ *resty.Client, r *resty.Response) (time.Duration, error) {
	if r == nil {
		return defaultRateLimitWait, nil
	}
	if seconds, err := strconv.Atoi(r.Header().Get("Retry-After")); err == nil && seconds > 0 {
		return min(time.Duration(seconds)*time.Second, maxRateLimitWait), nil
	}
	return defaultRateLimitWait, nil
}

func (c *Client) SetDebug(debug bool) *Client {
	c.Debug = debug

	return c
}

func (c *Client) Get(ctx context.Context, cr *ClientRequest) (interface{}, *Error) {
	hasResult := cr.Result != nil

	request := c.HttpClient.R().
		SetContext(ctx).
		SetError(&Error{}).
		SetQueryParams(cr.QueryParams)

	if hasResult {
		request.SetResult(cr.Result)
	}

	resp, err := request.
		Get(cr.Path)
	if err != nil {
		return nil, &Error{Err: err}
	}

	if c.Debug {
		fmt.Println("REQUEST: ", resp.Request.RawRequest)
		fmt.Println("RESPONSE: ", resp.String())
	}

	if resp.IsError() {
		err, ok := resp.Error().(*Error)
		if !ok {
			return nil, &ErrorTypeAssert
		}

		return nil, c.handleNonJSONErrors(resp, err)
	}

	if hasResult {
		return resp.Result(), nil
	}

	return resp.String(), nil
}

func (c *Client) Post(ctx context.Context, cr *ClientRequest) (interface{}, *Error) {
	resp, err := c.HttpClient.R().
		SetContext(ctx).
		SetError(&Error{}).
		SetResult(cr.Result).
		SetBody(cr.Body).
		Post(cr.Path)
	if err != nil {
		return nil, &Error{Err: err}
	}

	if c.Debug {
		fmt.Println("REQUEST: ", resp.Request.RawRequest)
		fmt.Println("RESPONSE: ", resp.String())
	}

	if resp.IsError() {
		err, ok := resp.Error().(*Error)
		if !ok {
			return nil, &ErrorTypeAssert
		}

		return nil, c.handleNonJSONErrors(resp, err)
	}

	return resp.Result(), nil
}

func (c *Client) PostWithoutResult(ctx context.Context, cr *ClientRequest) *Error {
	resp, err := c.HttpClient.R().
		SetContext(ctx).
		SetError(&Error{}).
		SetBody(cr.Body).
		Post(cr.Path)
	if err != nil {
		return &Error{Err: err}
	}

	if c.Debug {
		fmt.Println("REQUEST: ", resp.Request.RawRequest)
		fmt.Println("RESPONSE: ", resp.String())
	}

	if resp.IsError() {
		err, ok := resp.Error().(*Error)
		if !ok {
			return &ErrorTypeAssert
		}

		return err
	}

	return nil
}

func (c *Client) PostWithoutBody(ctx context.Context, cr *ClientRequest) (interface{}, *Error) {
	resp, err := c.HttpClient.R().
		SetContext(ctx).
		SetError(&Error{}).
		Post(cr.Path)
	if err != nil {
		return nil, &Error{Err: err}
	}

	if c.Debug {
		fmt.Println("REQUEST: ", resp.Request.RawRequest)
		fmt.Println("RESPONSE: ", resp.String())
	}

	if resp.IsError() {
		err, ok := resp.Error().(*Error)
		if !ok {
			return nil, &ErrorTypeAssert
		}

		return nil, c.handleNonJSONErrors(resp, err)
	}

	return resp.Result(), nil
}

func (c *Client) Put(ctx context.Context, cr *ClientRequest) (interface{}, *Error) {
	resp, err := c.HttpClient.R().
		SetContext(ctx).
		SetError(&Error{}).
		SetResult(cr.Result).
		SetBody(cr.Body).
		Put(cr.Path)
	if err != nil {
		return nil, &Error{Err: err}
	}

	if c.Debug {
		fmt.Println("REQUEST: ", resp.Request.RawRequest)
		fmt.Println("RESPONSE: ", resp.String())
	}

	if resp.IsError() {
		err, ok := resp.Error().(*Error)
		if !ok {
			return nil, &ErrorTypeAssert
		}

		return nil, c.handleNonJSONErrors(resp, err)
	}

	return resp.Result(), nil
}

func (c *Client) Delete(ctx context.Context, cr *ClientRequest) (interface{}, *Error) {
	resp, err := c.HttpClient.R().
		SetContext(ctx).
		SetError(&Error{}).
		SetResult(cr.Result).
		SetBody(cr.Body).
		SetQueryParams(cr.QueryParams).
		Delete(cr.Path)
	if err != nil {
		return nil, &Error{Err: err}
	}

	if c.Debug {
		fmt.Println("REQUEST: ", resp.Request.RawRequest)
		fmt.Println("RESPONSE: ", resp.String())
	}

	if resp.IsError() {
		err, ok := resp.Error().(*Error)
		if !ok {
			return nil, &ErrorTypeAssert
		}

		return nil, c.handleNonJSONErrors(resp, err)
	}

	return resp.Result(), nil
}

func (c *Client) handleNonJSONErrors(resp *resty.Response, err *Error) *Error {
	// there are some cases where the API doesn't respond with a proper json response:
	// - if nginx fails the cert validation (or other errors like 502, max body exceeded etc.)
	// - if foundry is in dev mode and throws an error
	// in both cases, we get a html response. try to make sense out of this error.
	if err.HTTPStatusCode == 0 {
		return &Error{
			HTTPStatusCode: resp.StatusCode(),
			Message:        resp.Status(),
		}
	}

	return err
}

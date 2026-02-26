package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultHTTPHost = "http://localhost:8080"
)

var (
	Headers = map[string]string{"content-type": "application/json; charset=utf-8"}
)

type Client struct {
	httpc *resty.Client
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// BloomAddRequest represents the request body for adding keys to bloom filter
type BloomAddRequest struct {
	Key  string   `json:"key,omitempty"`
	Keys []string `json:"keys,omitempty"`
}

// BloomCheckResponse represents the response from checking a key
type BloomCheckResponse struct {
	Key    string `json:"key"`
	Exists bool   `json:"exists"`
}

func NewClientWithResty(httpc *resty.Client, opts ...ClientOpt) (*Client, error) {
	c := &Client{
		httpc: httpc,
	}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	if c.httpc.HostURL == "" {
		c.httpc.SetBaseURL(DefaultHTTPHost)
	}
	for name, val := range Headers {
		c.httpc.SetHeader(name, val)
	}

	return c, nil
}

func NewClient(opts ...ClientOpt) (*Client, error) {
	return NewClientWithResty(resty.New(), opts...)
}

// AddToBloom adds a single key to the bloom filter
func (c *Client) AddToBloom(ctx context.Context, key string) (Response, error) {
	req := c.httpc.R()

	body := BloomAddRequest{Key: key}

	resp, err := req.SetContext(ctx).SetBody(body).Post("/bloom/add")
	if err != nil {
		return Response{}, fmt.Errorf("writer_lambda: %w", err)
	}

	result := Response{
		StatusCode: resp.StatusCode(),
		Body:       string(resp.Body()),
		Headers:    Headers,
	}

	return result, nil
}

// AddMultipleToBloom adds multiple keys to the bloom filter
func (c *Client) AddMultipleToBloom(ctx context.Context, keys []string) (Response, error) {
	req := c.httpc.R()

	body := BloomAddRequest{Keys: keys}

	resp, err := req.SetContext(ctx).SetBody(body).Post("/bloom/add")
	if err != nil {
		return Response{}, fmt.Errorf("writer_lambda: %w", err)
	}

	result := Response{
		StatusCode: resp.StatusCode(),
		Body:       string(resp.Body()),
		Headers:    Headers,
	}

	return result, nil
}

// CheckBloom checks if a key exists in the bloom filter
func (c *Client) CheckBloom(ctx context.Context, key string) (BloomCheckResponse, error) {
	req := c.httpc.R()

	var checkResp BloomCheckResponse
	resp, err := req.SetContext(ctx).SetResult(&checkResp).Get("/bloom/check/" + key)
	if err != nil {
		return BloomCheckResponse{}, fmt.Errorf("writer_lambda: %w", err)
	}

	if resp.StatusCode() != 200 {
		return BloomCheckResponse{}, fmt.Errorf("writer_lambda: unexpected status %d", resp.StatusCode())
	}

	return checkResp, nil
}

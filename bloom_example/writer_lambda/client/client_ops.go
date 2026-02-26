package client

import (
	"errors"
)

// ClientOpt is a function that sets client behavior or basic data.
// Example:
//
//	c, err := NewClient(
//	    AsService("writer_lambda"),
//	    WithHTTPHost("http://cache:8080"),
//	)
type ClientOpt func(c *Client) error

// AsService supplies service name for http client for debugging.
func AsService(name string) ClientOpt {
	return func(c *Client) error {
		if name == "" {
			return errors.New("writer_lambda: service name cannot be empty")
		}
		c.httpc.SetQueryParam("service", name)
		return nil
	}
}

// WithHTTPHost rewrites default HTTP host in client.
func WithHTTPHost(host string) ClientOpt {
	return func(c *Client) error {
		c.httpc.SetBaseURL(host)
		return nil
	}
}

// WithDebug enables debug resty output
func WithDebug(enable bool) ClientOpt {
	return func(c *Client) error {
		c.httpc.SetDebug(enable)
		return nil
	}
}

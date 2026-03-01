package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
	headers map[string]string
}

type Option func(*Client)

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = timeout
	}
}

func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

func (c *Client) Do(
	ctx context.Context,
	method,
	endpoint string,
	token string,
	body any,
) ([]byte, int, error) {

	url := fmt.Sprintf("%s/%s", c.baseURL, strings.TrimLeft(endpoint, "/"))

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

func (c *Client) Get(ctx context.Context, endpoint, token string) ([]byte, int, error) {
	return c.Do(ctx, http.MethodGet, endpoint, token, nil)
}

func (c *Client) Post(ctx context.Context, endpoint, token string, body any) ([]byte, int, error) {
	return c.Do(ctx, http.MethodPost, endpoint, token, body)
}

func (c *Client) Put(ctx context.Context, endpoint, token string, body any) ([]byte, int, error) {
	return c.Do(ctx, http.MethodPut, endpoint, token, body)
}

func (c *Client) Patch(ctx context.Context, endpoint, token string, body any) ([]byte, int, error) {
	return c.Do(ctx, http.MethodPatch, endpoint, token, body)
}

func (c *Client) Delete(ctx context.Context, endpoint, token string, body any) ([]byte, int, error) {
	return c.Do(ctx, http.MethodDelete, endpoint, token, body)
}

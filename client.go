// Package uptime provides a Go client for the Tickstem uptime monitoring API.
//
// Usage:
//
//	client := uptime.New(os.Getenv("TICKSTEM_API_KEY"))
//
//	monitor, err := client.Create(ctx, uptime.CreateParams{
//	    Name: "Production API",
//	    URL:  "https://api.yourapp.com/health",
//	})
//
//	checks, err := client.Checks(ctx, monitor.ID, uptime.ChecksParams{Limit: 50})
package uptime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.tickstem.dev/v1"

// Client is a Tickstem uptime monitoring API client. Create one with New.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL. Useful for local testing.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New creates a Tickstem uptime monitoring client. The client is safe for
// concurrent use and should be created once and reused.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// List returns all monitors for the authenticated account.
func (c *Client) List(ctx context.Context) ([]*Monitor, error) {
	var resp listMonitorsResponse
	if err := c.do(ctx, http.MethodGet, "/monitors", nil, &resp); err != nil {
		return nil, err
	}
	if resp.Monitors == nil {
		return []*Monitor{}, nil
	}
	return resp.Monitors, nil
}

// Get returns a single monitor by ID.
func (c *Client) Get(ctx context.Context, id string) (*Monitor, error) {
	var m Monitor
	if err := c.do(ctx, http.MethodGet, "/monitors/"+id, nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Create registers a new uptime monitor.
func (c *Client) Create(ctx context.Context, params CreateParams) (*Monitor, error) {
	var m Monitor
	if err := c.do(ctx, http.MethodPost, "/monitors", params, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Update changes one or more fields on an existing monitor.
func (c *Client) Update(ctx context.Context, id string, params UpdateParams) (*Monitor, error) {
	var m Monitor
	if err := c.do(ctx, http.MethodPatch, "/monitors/"+id, params, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Pause suspends checks for a monitor without deleting it.
func (c *Client) Pause(ctx context.Context, id string) (*Monitor, error) {
	var m Monitor
	if err := c.do(ctx, http.MethodPatch, "/monitors/"+id+"/pause", nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Resume re-enables checks for a paused monitor.
func (c *Client) Resume(ctx context.Context, id string) (*Monitor, error) {
	var m Monitor
	if err := c.do(ctx, http.MethodPatch, "/monitors/"+id+"/resume", nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Delete permanently removes a monitor and all its check history.
func (c *Client) Delete(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/monitors/"+id, nil, nil)
}

// Checks returns recent check results for a monitor in reverse chronological order.
func (c *Client) Checks(ctx context.Context, monitorID string, params ChecksParams) ([]*MonitorCheck, error) {
	path := fmt.Sprintf("/monitors/%s/checks?limit=%d&offset=%d",
		monitorID,
		checksLimitOrDefault(params.Limit),
		params.Offset,
	)
	var resp listChecksResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	if resp.Checks == nil {
		return []*MonitorCheck{}, nil
	}
	return resp.Checks, nil
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	req, err := c.buildRequest(ctx, method, path, body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tickstem/uptime: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("tickstem/uptime: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("tickstem/uptime: decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) buildRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("tickstem/uptime: encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("tickstem/uptime: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "tickstem-go/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func checksLimitOrDefault(limit int) int {
	if limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}

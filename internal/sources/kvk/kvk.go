package kvk

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client for KVK (Netherlands)
// API: https://api.kvk.nl/api/v1
// Auth: OAuth2 bearer token required
type Client struct {
	client   *http.Client
	token    string
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
		token:  "", // OAuth2 token
	}
}

func (c *Client) Name() string { return "kvk" }

func (c *Client) FetchCompanies(_ context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("kvk: requires OAuth2 bearer token — set token field")
}

func (c *Client) FetchFinancials(_ context.Context, _ string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("kvk: financial data requires elevated OAuth2 permissions")
}

func (c *Client) FetchProfile(_ context.Context, _ string) (*sources.Profile, error) {
	return nil, fmt.Errorf("kvk: profile requires OAuth2 authentication")
}

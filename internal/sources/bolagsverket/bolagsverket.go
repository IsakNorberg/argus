package bolagsverket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client for Bolagsverket (Sweden)
// API: https://data.bolagsverket.se/api/v1
// Auth: API key required (header: X-Api-Key)
type Client struct {
	client   *http.Client
	apiKey   string
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
		apiKey: "", // TODO: Set via env or config
	}
}

func (c *Client) Name() string { return "bolagsverket" }

func (c *Client) FetchCompanies(_ context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("bolagsverket: requires API key — set apiKey field or env var")
}

func (c *Client) FetchFinancials(_ context.Context, _ string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("bolagsverket: financial data behind paywall (UC/Bisnode)")
}

func (c *Client) FetchProfile(_ context.Context, _ string) (*sources.Profile, error) {
	return nil, fmt.Errorf("bolagsverket: profile requires API key")
}

package ares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client for ARES (Czech Republic)
// API: https://ares.gov.cz
// Auth: None, but uses complex nested JSON responses
type Client struct {
	client   *http.Client
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "ares" }

func (c *Client) FetchCompanies(_ context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("ares: complex API — no simple company list endpoint available")
}

func (c *Client) FetchFinancials(_ context.Context, _ string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("ares: financial data requires nested API calls")
}

func (c *Client) FetchProfile(_ context.Context, _ string) (*sources.Profile, error) {
	return nil, fmt.Errorf("ares: profile requires parsing complex JSON structure")
}

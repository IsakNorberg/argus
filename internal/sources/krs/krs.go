package krs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client for KRS/EMS (Poland)
// API: https://ems.ms.gov.pl
// Auth: None, but endpoint returns 302 redirect to login
type Client struct {
	client   *http.Client
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "krs" }

func (c *Client) FetchCompanies(_ context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("krs: EMS endpoint returns 302 redirect to login — no public API")
}

func (c *Client) FetchFinancials(_ context.Context, _ string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("krs: no public API for financial data")
}

func (c *Client) FetchProfile(_ context.Context, _ string) (*sources.Profile, error) {
	return nil, fmt.Errorf("krs: requires authenticated access")
}

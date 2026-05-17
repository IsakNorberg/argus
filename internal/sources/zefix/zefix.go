package zefix

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client for Zefix (Switzerland)
// API: https://www.zefix.admin.ch
// Auth: None, but uses complex SPA frontend — no documented REST API
type Client struct {
	client   *http.Client
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "zefix" }

func (c *Client) FetchCompanies(_ context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("zefix: no REST API documented — SPA frontend only")
}

func (c *Client) FetchFinancials(_ context.Context, _ string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("zefix: financial data not available via public interface")
}

func (c *Client) FetchProfile(_ context.Context, _ string) (*sources.Profile, error) {
	return nil, fmt.Errorf("zefix: profile requires SPA rendering")
}

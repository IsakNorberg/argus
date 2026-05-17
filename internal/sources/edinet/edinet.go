package edinet

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client for EDINET (Japan)
// API: https://disclosure.edinet-fsa.go.jp/api/v1/
// Auth: None, but all docs are HTML/PDF and require parsing
type Client struct {
	client   *http.Client
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "edinet" }

func (c *Client) FetchCompanies(_ context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("edinet: company list requires parsing XBRL/XML documents")
}

func (c *Client) FetchFinancials(_ context.Context, _ string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("edinet: financials in XBRL format — parser needed")
}

func (c *Client) FetchProfile(_ context.Context, _ string) (*sources.Profile, error) {
	return nil, fmt.Errorf("edinet: profile requires parsing HTML pages")
}

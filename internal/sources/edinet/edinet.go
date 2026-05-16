package edinet

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client hämtar data från EDINET (Japan).
type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://disclosure2.edinet-fsa.go.jp",
	}
}

func (c *Client) Name() string { return "edinet" }

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: EDINET har XBRL-filer för alla japanska bolag
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: Hämta XBRL från EDINET
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

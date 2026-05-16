package zefix

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client hämtar data från Zefix (Schweiz).
type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://www.zefix.ch",
	}
}

func (c *Client) Name() string { return "zefix" }

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: Zefix har öppet REST API
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

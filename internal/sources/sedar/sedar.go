package sedar

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client hämtar data från SEDAR+ (Kanada).
type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://api.sedarplus.ca",
	}
}

func (c *Client) Name() string { return "sedar" }

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: SEDAR+ är Kanadas motsvarighet till SEC EDGAR
	// Hämta filing-arkivet från https://sedarplus.ca
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: Kanadensiska bolag rapporterar i XBRL
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

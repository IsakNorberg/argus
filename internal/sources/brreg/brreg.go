package brreg

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client hämtar data från Brønnøysundregistrene (Norge).
type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://data.brreg.no",
	}
}

func (c *Client) Name() string { return "brreg" }

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: Implementera
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: Implementera
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	// TODO: Implementera
	return nil, fmt.Errorf("inte implementerad ännen")
}

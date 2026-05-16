package ares

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client hämtar data från ARES (Tjeckien).
type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://ares.gov.cz",
	}
}

func (c *Client) Name() string { return "ares" }

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

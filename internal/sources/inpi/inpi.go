package inpi

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client hämtar data från INPI / BODACC (Frankrike).
type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://data.inpi.fr",
	}
}

func (c *Client) Name() string { return "inpi" }

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

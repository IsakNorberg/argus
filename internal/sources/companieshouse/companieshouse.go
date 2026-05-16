package companieshouse

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://api.company-information.service.gov.uk",
	}
}

func (c *Client) Name() string { return "companieshouse" }

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

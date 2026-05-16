package cvr

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
		baseURL: "https://datacvr.virk.dk/api",
	}
}

func (c *Client) Name() string { return "cvr" }

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

package sec

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

type SECClient struct {
	baseURL string
}

func New() *SECClient {
	return &SECClient{
		baseURL: "https://data.sec.gov",
	}
}

func (c *SECClient) Name() string { return "sec" }

func (c *SECClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("ej implementerad ännu")
}

func (c *SECClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("ej implementerad ännu")
}

func (c *SECClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

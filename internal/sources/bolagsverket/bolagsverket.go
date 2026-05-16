package bolagsverket

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

type BVClient struct {
	baseURL string
}

func New() *BVClient {
	return &BVClient{
		baseURL: "https://bolagsverket.se",
	}
}

func (c *BVClient) Name() string { return "bolagsverket" }

func (c *BVClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

func (c *BVClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

func (c *BVClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("ej implementerad ännen")
}

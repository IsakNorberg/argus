package morningstar

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// MorningstarClient hämtar data från Morningstar.
type MorningstarClient struct {
	baseURL string
}

func New() *MorningstarClient {
	return &MorningstarClient{
		baseURL: "https://api-global.morningstar.com",
	}
}

func (c *MorningstarClient) Name() string { return "morningstar" }

func (c *MorningstarClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *MorningstarClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: /sal-service/1/stock/v2/KeyRatios/{symbol}
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *MorningstarClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

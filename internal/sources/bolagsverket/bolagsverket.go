package bolagsverket

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// BVClient hämtar data från Bolagsverket API.
type BVClient struct {
	apiKey string
}

func New(apiKey string) *BVClient {
	return &BVClient{apiKey: apiKey}
}

func (c *BVClient) Name() string { return "bolagsverket" }

func (c *BVClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: Bolagsverket har REST API men kräver OAuth/API-nyckel
	// https://bolagsverket.se/datamangder/oppnadata
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *BVClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: Årsredovisningar finns, men många är PDF/digitala
	return nil, fmt.Errorf("inte implementerad ännen")
}

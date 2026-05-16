package standardizer

import (
	"github.com/IsakNorberg/argus/internal/sources"
	"github.com/IsakNorberg/argus/pkg/models"
)

// Standardizer konverterar käll-specifik data till gemensamt format.
type Standardizer struct{}

func New() *Standardizer { return &Standardizer{} }

// StandardizeCompany omvandlar en Source.Company till en models.Company.
func (s *Standardizer) StandardizeCompany(c sources.Company, sourceName string) *models.Company {
	mc := &models.Company{
		Name:    c.Name,
		Ticker:  c.Ticker,
		Country: c.Country,
		Exchange: c.Exchange,
		Industry: c.Industry,
	}

	if sourceName == "sec" {
		mc.CIK = c.ExternalID
	} else if sourceName == "bolagsverket" {
		mc.OrgNumber = c.ExternalID
	}

	return mc
}

// StandardizeFinancials omvandlar Source.Financials till models.Financials.
func (s *Standardizer) StandardizeFinancials(f sources.Financials) *models.Financials {
	return &models.Financials{
		Currency:     f.Currency,
		Revenue:      f.Revenue,
		NetIncome:    f.NetIncome,
		TotalAssets:  f.TotalAssets,
		Source:       f.Source,
		Period:       f.Period,
	}
}

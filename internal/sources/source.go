package sources

import "context"

// Source är gränssnittet för alla datakällor.
// Varje källa (SEC, Bolagsverket, etc) implementerar detta.
type Source interface {
	// Name returnerar källans namn, t.ex. "sec"
	Name() string

	// FetchCompanies hämtar bolagslista
	FetchCompanies(ctx context.Context) ([]Company, error)

	// FetchFinancials hämtar finansiell data för ett bolag
	FetchFinancials(ctx context.Context, companyID string) ([]Financials, error)
}

// Company – gemensamt format för bolag från alla källor.
type Company struct {
	ID         string
	Name       string
	Ticker     string
	Country    string
	Industry   string
	Exchange   string
	ExternalID string // Källans eget ID (CIK, orgnummer, etc)
}

// Financials – gemensamt format för finansiell data.
type Financials struct {
	CompanyID  string
	Period     string
	Currency   string
	Revenue    float64
	NetIncome  float64
	TotalAssets float64
	Source     string
	ReportDate string
	FilingDate string
	RawData    map[string]any // Behåll originalet för debugging
}

package sources

import "context"

// Source är gränssnittet för alla datakällor.
type Source interface {
	// Name returnerar källans namn, t.ex. "sec", "eodhd"
	Name() string

	// FetchCompanies hämtar bolagslista
	FetchCompanies(ctx context.Context) ([]Company, error)

	// FetchFinancials hämtar finansiell data för ett bolag
	FetchFinancials(ctx context.Context, companyID string) ([]Financials, error)

	// FetchProfile hämtar bolagsprofil (beskrivning, sektor, etc)
	FetchProfile(ctx context.Context, companyID string) (*Profile, error)
}

// Company – gemensamt format för bolag från alla källor.
type Company struct {
	ID         string // Intern argus ID
	Name       string
	Ticker     string
	Country    string
	Industry   string
	Exchange   string
	ExternalID string // Källans eget ID (CIK, orgnr, etc)
}

// Financials – gemensamt format för finansiell data.
type Financials struct {
	CompanyID   string
	Period      string    // "2024Q4", "2024FY"
	Currency    string
	Revenue     float64
	NetIncome   float64
	TotalAssets float64
	TotalDebt   float64
	Equity      float64
	EPS         float64
	EBITDA      float64
	Source      string
	ReportDate  string
	FilingDate  string
	RawData     map[string]any
}

// Profile – bolagsprofil oberoende av källa.
type Profile struct {
	CompanyID   string
	Description string
	Sector      string
	Industry    string
	Country     string
	Exchange    string
	Employees   int
	Website     string
	CEO         string
}

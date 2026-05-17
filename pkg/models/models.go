package models

import "time"

// Company representerar ett bolag med gemensamma fält oavsett källa.
type Company struct {
	ID           int64     `json:"id"`
	ArgusID      string    `json:"argus_id"`      // Vårt unika ID
	Name         string    `json:"name"`
	Ticker       string    `json:"ticker"`
	CIK          string    `json:"cik"`            // SEC
	OrgNumber    string    `json:"org_number"`     // Bolagsverket
	ISIN         string    `json:"isin"`           // Internationellt
	Country      string    `json:"country"`
	Industry     string    `json:"industry"`
	Sector       string    `json:"sector"`
	Exchange     string    `json:"exchange"`
	Description  string    `json:"description"`
	Website      string    `json:"website"`
	Employees    int       `json:"employees"`
	CEO          string    `json:"ceo"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Financials är standardiserade siffror från alla källor.
type Financials struct {
	ID           int64     `json:"id"`
	ArgusID      string    `json:"argus_id"`
	CompanyID    int64     `json:"company_id"`
	Period       string    `json:"period"`          // "2024Q4", "2024FY"
	Currency     string    `json:"currency"`
	Revenue      float64   `json:"revenue"`
	NetIncome    float64   `json:"net_income"`
	TotalAssets  float64   `json:"total_assets"`
	TotalDebt    float64   `json:"total_debt"`
	Equity       float64   `json:"equity"`
	EPS          float64   `json:"eps"`
	EBITDA       float64   `json:"ebitda"`
	FreeCashFlow float64   `json:"free_cash_flow"`
	ReportDate   time.Time `json:"report_date"`
	FilingDate   time.Time `json:"filing_date"`
	Source       string    `json:"source"`          // "sec", "eodhd", etc
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SourceMetadata spårar vad vi har hämtat från varje källa.
type SourceMetadata struct {
	ID          int64     `json:"id"`
	ArgusID     string    `json:"argus_id"`
	Source      string    `json:"source"`
	LastFetched time.Time `json:"last_fetched"`
	RecordCount int       `json:"record_count"`
	ErrorCount  int       `json:"error_count"`
	Status      string    `json:"status"` // "ok", "error", "partial"
	CreatedAt   time.Time `json:"created_at"`
}

// PriceQuote är en dagprisnotering från en kurskälla.
type PriceQuote struct {
	ID            int64     `json:"id"`
	ArgusID       string    `json:"argus_id"`
	Ticker        string    `json:"ticker"`
	Price         float64   `json:"price"`
	Currency      string    `json:"currency"`
	MarketCap     int64     `json:"market_cap"`
	Volume        int64     `json:"volume"`
	ChangePercent float64   `json:"change_percent"`
	Source        string    `json:"source"` // "yahoo", "eodhd", etc
	QuoteDate     time.Time `json:"quote_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

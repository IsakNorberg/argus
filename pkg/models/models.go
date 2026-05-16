package models

import "time"

// Company representerar ett bolag med gemensamma fält oavsett källa.
type Company struct {
	CIK            string    `json:"cik"`             // SEC CIK (tom om ej SEC)
	OrgNumber      string    `json:"org_number"`      // Bolagsverket organisationsnummer
	Name           string    `json:"name"`
	Ticker         string    `json:"ticker"`          // Börsförkortning
	Country        string    `json:"country"`
	Industry       string    `json:"industry"`
	Sector         string    `json:"sector"`
	Exchange       string    `json:"exchange"`        // T.ex. "NASDAQ", "SSE"
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Financials är standardiserade siffror från alla källor.
type Financials struct {
	CompanyID      int64     `json:"company_id"`
	Period         string    `json:"period"`          // "2024Q4", "2024FY"
	Currency       string    `json:"currency"`
	Revenue        float64   `json:"revenue"`
	NetIncome      float64   `json:"net_income"`
	TotalAssets    float64   `json:"total_assets"`
	TotalDebt      float64   `json:"total_debt"`
	Equity         float64   `json:"equity"`
	EPS            float64   `json:"eps"`
	EBITDA         float64   `json:"ebitda"`
	FreeCashFlow   float64   `json:"free_cash_flow"`
	ReportDate     time.Time `json:"report_date"`
	FilingDate     time.Time `json:"filing_date"`
	Source         string    `json:"source"`          // "sec", "bolagsverket", etc.
	CreatedAt      time.Time `json:"created_at"`
}

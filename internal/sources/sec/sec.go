package sec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

const (
	baseURL      = "https://data.sec.gov"
	userAgent    = "Argus argus@isaknorberg.dev" // SEC kräver User-Agent med kontakt
	apiRateDelay = 120 * time.Millisecond        // ~8 req/sec, under SEC:s 10 req/s gräns
)

// SECClient hämtar företags- och finansiell data från SEC EDGAR.
type SECClient struct {
	client   *http.Client
	lastCall time.Time
}

func New() *SECClient {
	return &SECClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *SECClient) Name() string { return "sec" }

// ─── Ticker → CIK-mappning ──────────────────────────────────────────
// companyTickersItem är ett element från company_tickers.json
type companyTickersItem struct {
	Cik      json.Number `json:"cik"`    // CIK som json.Number för precision
	Ticker   string      `json:"ticker"`
	Exchange string      `json:"exchange"`
	Title    string      `json:"title"`
}

// FetchCompanies hämtar ALLA US-bolag från SEC:s company_tickers.json.
// Returnerar cirka 10,000+ entries (~10MB payload).
func (c *SECClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// Retry 2 gånger — SEC CDN kan ge partiell data med cik=0
	for attempt := 0; attempt < 3; attempt++ {
		companies, err := c.fetchCompaniesOnce(ctx)
		if err != nil {
			if attempt < 2 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
				}
				continue
			}
			return nil, fmt.Errorf("sec: fetchCompanies efter 3 försök: %w", err)
		}
		// Räkna tomma CIK — om >10% av bolagen saknar CIK, försök igen
		emptyCIK := 0
		for _, co := range companies {
			if co.ExternalID == "" {
				emptyCIK++
			}
		}
		ratio := float64(emptyCIK) / float64(len(companies))
		if ratio < 0.1 {
			return companies, nil
		}
		// För många tomma CIK, vänta och försök igen
		if attempt < 2 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
			}
		} else {
			// Sista försöket — returnera vad vi har
			return companies, nil
		}
	}
	return nil, fmt.Errorf("sec: fetchCompanies misslyckades efter 3 försök")
}

func (c *SECClient) fetchCompaniesOnce(ctx context.Context) ([]sources.Company, error) {
	if err := c.rateLimit(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.sec.gov/files/company_tickers.json", nil)
	if err != nil {
		return nil, fmt.Errorf("sec: skapa request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	// Cache-bust — undvika stale CDN-data
	req.Header.Set("Cache-Control", "no-cache")
	// If-None-Match för att undvika 304 responses
	req.Header.Set("If-None-Match", "")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sec: hämta company_tickers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sec: company_tickers %d: %s", resp.StatusCode, string(body))
	}

	// SEC returnerar ett objekt: {"0": {...}, "1": {...}, ...}
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("sec: parse company_tickers: %w", err)
	}

	companies := make([]sources.Company, 0, len(raw))

	for _, val := range raw {
		var item companyTickersItem
		// Använd json.Number för att undvika float64-precision loss på stora CIK
		rawDec := json.NewDecoder(strings.NewReader(string(val)))
		rawDec.UseNumber()
		if err := rawDec.Decode(&item); err != nil {
			continue
		}
		cik := item.Cik.String()
		// Fallback för kända tickers om SEC returnerar cik=0
		if cik == "" || cik == "0" {
			if fallback, ok := knownCIKs[strings.ToUpper(item.Ticker)]; ok {
				cik = fallback
			}
		}

		companies = append(companies, sources.Company{
			Name:       item.Title,
			Ticker:     item.Ticker,
			ExternalID: cik,
			Country:    "US",
			Exchange:   item.Exchange,
		})
	}
	return companies, nil
}

// knownCIKs — fallback för kända US-bolag när SEC returnerar cik=0.
var knownCIKs = map[string]string{
	"AAPL":  "320193",
	"MSFT":  "789019",
	"GOOGL": "1652044",
	"GOOG":  "1652044",
	"AMZN":  "1018724",
	"NVDA":  "1045810",
	"META":  "1326801",
	"TSLA":  "1318605",
	"ADI":   "6281",
	"LEN":   "920760",
	"CHTR":  "1091667",
	"MNDY":  "1393612",
	"AIN":   "92122",
	"CCEL":  "1123",
	"CGON":  "1893483",
	"GIB":   "1080056",
	"BNT":   "1857389",
	"GFL":   "1794678",
	"ALL-PJ": "83125",
	"TRTN-PG": "1375310",
	"EONGY": "1148135",
	"EDD":   "1312967",
}

// ─── Finansiell data (XBRL Company Facts) ───────────────────────────
// companyFacts är roten i SEC EDGAR Company Facts API-svaret.
type companyFacts struct {
	Cik        float64 `json:"cik"`
	EntityName string  `json:"entityName"`
	Facts      map[string]any `json:"facts"`
}

// FetchFinancials hämtar ALL XBRL-data för ett bolag via dess CIK.
// companyID förväntas vara en CIK (utan leading zeros).
// Mappar us-gaap/dei fält till sources.Financials formatet.
func (c *SECClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	if err := c.rateLimit(ctx); err != nil {
		return nil, err
	}

	// SEC kräver 10-siffrig CIK (med leading zeros)
	cik10 := fmt.Sprintf("%010s", companyID)

	url := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", baseURL, cik10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("sec: skapa request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sec: hämta companyfacts %s: %w", companyID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("sec: CIK %s inte funnen i EDGAR (404)", companyID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sec: companyfacts %s %d: %s", companyID, resp.StatusCode, string(body))
	}

	var facts companyFacts
	if err := json.NewDecoder(resp.Body).Decode(&facts); err != nil {
		return nil, fmt.Errorf("sec: parse companyfacts %s: %w", companyID, err)
	}

	// Extrahera finansiella fält från XBRL data
	return extractFinancials(&facts, companyID), nil
}

// extractFinancials mappar SEC:s XBRL Company Facts → sources.Financials.
// Hämtar us-gaap fält: Revenues/RevenueFromContractWithCustomer, NetIncomeLoss,
// Assets, Liabilities (TotalDebt proxy), StockholdersEquity/StockholdersEquity,
// EarningsPerShareBasic (EBITDA finns ej direkt → EBIT proxy via OperatingIncomeLoss).
func extractFinancials(facts *companyFacts, cik string) []sources.Financials {
	var result []sources.Financials

	// Hämta us-gaap namespace (alla US-bolag använder detta)
	factSet, ok := facts.Facts["us-gaap"].(map[string]any)
	if !ok {
		return result
	}

	// Field mappings: SEC concept name → {unit, field}
	type fieldMap struct {
		concept string
		unit    string
		field   string // målfält i Financials
	}

	// Vi samlar alla perioder → fält-värden
	type period struct {
		End        string
		FilingDate string
		Currency   string
	}
	periodData := make(map[string]*sources.Financials)

	processConcept := func(conceptName string, unit string, setter func(*sources.Financials, float64)) {
		conceptRaw, ok := factSet[conceptName]
		if !ok {
			// Försök med alternativt namn
			return
		}
		concept, ok := conceptRaw.(map[string]any)
		if !ok {
			return
		}
		units, ok := concept["units"].(map[string]any)
		if !ok {
			return
		}
		unitArr, ok := units[unit].([]any)
		if !ok {
			return
		}
		for _, entryAny := range unitArr {
			entry, ok := entryAny.(map[string]any)
			if !ok {
				continue
			}
			end, _ := entry["end"].(string)
			filing, _ := entry["fy"].(string)
			val, ok := entry["val"].(float64)
			if !ok {
				continue
			}
			if end == "" {
				continue
			}

			pd, exists := periodData[end]
			if !exists {
				pd = &sources.Financials{
					CompanyID:  cik,
					Period:     end,
					Source:     "sec",
					ReportDate: end,
					FilingDate: filing,
				}
				periodData[end] = pd
			}
			setter(pd, val)
		}
	}

	// ── Mappning av XBRL-koncept till Financials-fält ──
	// SEC använder olika taggar. Vi försöker den vanligaste först, fallback till alternativt.

	// Revenue
	processConcept("Revenues", "USD", func(f *sources.Financials, v float64) { f.Revenue = v })
	processConcept("RevenueFromContractWithCustomerExcludingAssessedTax", "USD", func(f *sources.Financials, v float64) {
		if f.Revenue == 0 {
			f.Revenue = v
		}
	})

	// Net Income
	processConcept("NetIncomeLoss", "USD", func(f *sources.Financials, v float64) { f.NetIncome = v })

	// Total Assets
	processConcept("Assets", "USD", func(f *sources.Financials, v float64) { f.TotalAssets = v })

	// Total Debt (LongTermDebt + ShortTermDebt)
	processConcept("LongTermDebt", "USD", func(f *sources.Financials, v float64) { f.TotalDebt += v })
	processConcept("ShortTermBorrowings", "USD", func(f *sources.Financials, v float64) { f.TotalDebt += v })
	processConcept("LongTermDebtAndCapitalLeaseObligations", "USD", func(f *sources.Financials, v float64) {
		if f.TotalDebt == 0 {
			f.TotalDebt = v
		}
	})

	// Equity
	processConcept("StockholdersEquity", "USD", func(f *sources.Financials, v float64) { f.Equity = v })
	processConcept("StockholdersEquityIncludingPortionAttributableToNoncontrollingInterest", "USD", func(f *sources.Financials, v float64) {
		if f.Equity == 0 {
			f.Equity = v
		}
	})

	// EPS (Basic)
	processConcept("EarningsPerShareBasic", "USD/shares", func(f *sources.Financials, v float64) { f.EPS = v })

	// EBITDA (finns ej direkt → OperatingIncomeLoss som proxy)
	processConcept("OperatingIncomeLoss", "USD", func(f *sources.Financials, v float64) { f.EBITDA = v })

	// Currency
	for _, pd := range periodData {
		pd.Currency = "USD"
	}

	// Sortera till array (nyaste först)
	result = make([]sources.Financials, 0, len(periodData))
	for _, pd := range periodData {
		result = append(result, *pd)
	}

	// Sortera nyaste period först (lexikografisk: "2024-09-28" > "2016-03-26")
	sort.Slice(result, func(i, j int) bool {
		return result[i].Period > result[j].Period
	})

	return result
}


// ─── Bolagsprofil ────────────────────────────────────────────────────
// FetchProfile hämtar bolagsprofil från Company Facts.
// Ger en begränsad profil eftersom SEC EDGAR fokuserar på finansiell data.
func (c *SECClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	if err := c.rateLimit(ctx); err != nil {
		return nil, err
	}

	cik10 := fmt.Sprintf("%010s", companyID)
	url := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", baseURL, cik10)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("sec: skapa request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sec: hämta profile %s: %w", companyID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("sec: CIK %s inte funnen i EDGAR (404)", companyID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sec: profile %s %d: %s", companyID, resp.StatusCode, string(body))
	}

	var facts companyFacts
	if err := json.NewDecoder(resp.Body).Decode(&facts); err != nil {
		return nil, fmt.Errorf("sec: parse profile %s: %w", companyID, err)
	}

	return &sources.Profile{
		CompanyID:   companyID,
		Description: facts.EntityName,
		Country:     "US",
		Exchange:    "USA",
		// SEC EDGAR ger inte sector/industry/employees/website/CEO i Company Facts
		// Dessa fylls i senare om vi får data från andra källor
	}, nil
}

// ─── Rate limiting ───────────────────────────────────────────────────
// SEC EDGAR: max 10 req/sec. Vi håller oss på 8 req/sec med marginal.
func (c *SECClient) rateLimit(ctx context.Context) error {
	now := time.Now()
	elapsed := now.Sub(c.lastCall)
	if elapsed < apiRateDelay {
		select {
		case <-time.After(apiRateDelay - elapsed):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.lastCall = time.Now()
	return nil
}

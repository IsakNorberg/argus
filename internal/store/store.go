package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/standardizer"
	"github.com/IsakNorberg/argus/internal/sources/sec"
	"github.com/IsakNorberg/argus/pkg/models"
)

// SECStore hanterar SEC EDGAR → SQLite-workflow.
type SECStore struct {
	db  database.DB
	sec *sec.SECClient
	std *standardizer.Standardizer
}

func New(db database.DB, secClient *sec.SECClient) *SECStore {
	return &SECStore{
		db:  db,
		sec: secClient,
		std: standardizer.New(),
	}
}

type FetchStats struct {
	CompaniesFetched  int
	CompaniesStored   int
	FinancialsFetched int
	FinancialsStored  int
	Errors            int
	Elapsed           time.Duration
}

// IngestCompanies hämtar alla US-bolag från SEC och upsertar dem i SQLite.
// Använder batch-insert med transaktion för prestanda.
// Returnerar statistik + ticker→CIK-mappning för vidare bruk.
func (s *SECStore) IngestCompanies(ctx context.Context) (*FetchStats, map[string]string, error) {
	start := time.Now()
	stats := &FetchStats{}

	companies, err := s.sec.FetchCompanies(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("hämta bolag från SEC: %w", err)
	}
	stats.CompaniesFetched = len(companies)
	tickerToCIK := make(map[string]string, len(companies))

	// Batch-insert med transaktion, 500 i taget
	batchSize := 500
	for i := 0; i < len(companies); i += batchSize {
		batch := companies[i:]
		if len(batch) > batchSize {
			batch = batch[:batchSize]
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("begin tx batch %d: %w", i, err)
		}

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO companies (argus_id, name, ticker, cik, country, exchange)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(argus_id) DO UPDATE SET
				name = excluded.name,
				ticker = excluded.ticker,
				country = excluded.country,
				exchange = excluded.exchange,
				updated_at = CURRENT_TIMESTAMP
		`)
		if err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("prepare batch %d: %w", i, err)
		}

		for _, c := range batch {
			_ = s.std.StandardizeCompany(c, "sec")
			argusID := "C" + c.ExternalID

			_, err := stmt.ExecContext(ctx, argusID, c.Name, c.Ticker,
				c.ExternalID, c.Country, c.Exchange)
			if err != nil {
				stats.Errors++
				continue
			}
			stats.CompaniesStored++
			tickerToCIK[strings.ToUpper(c.Ticker)] = c.ExternalID
		}
		stmt.Close()

		if err := tx.Commit(); err != nil {
			return nil, nil, fmt.Errorf("commit batch %d: %w", i, err)
		}
	}

	stats.Elapsed = time.Since(start)
	return stats, tickerToCIK, nil
}

// GetCIKByTicker slår upp CIK från den lokala databasen.
func (s *SECStore) GetCIKByTicker(ctx context.Context, ticker string) (string, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))

	c, err := s.db.CompanyByTicker(ctx, ticker)
	if err != nil {
		return "", fmt.Errorf("hitta %s i databasen: %w", ticker, err)
	}
	return c.CIK, nil
}

// FetchAndSaveFinancials hämtar alla finansiella rapporter för ett bolag (via CIK)
// och upsertar dem i SQLite. Bolaget måste redan finnas i databasen.
func (s *SECStore) FetchAndSaveFinancials(ctx context.Context, cik string) (*FetchStats, error) {
	start := time.Now()
	stats := &FetchStats{}
	cik = strings.TrimSpace(cik)

	// 1. Hämta intern company_id
	company, err := s.db.CompanyByCIK(ctx, cik)
	if err != nil {
		return nil, fmt.Errorf("hitta company_id för CIK %s: %w", cik, err)
	}

	// 2. Hämta finansiell data från SEC
	financials, err := s.sec.FetchFinancials(ctx, cik)
	if err != nil {
		return nil, fmt.Errorf("hämta financials för CIK %s: %w", cik, err)
	}
	stats.FinancialsFetched = len(financials)

	// 3. Standardisera och upserta varje period
	for _, f := range financials {
		mf := s.std.StandardizeFinancials(f)
		mf.CompanyID = company.ID
		mf.Source = "sec"
		mf.ArgusID = fmt.Sprintf("F-C%s-%s", cik, f.Period)

		if err := s.db.UpsertFinancials(ctx, mf); err != nil {
			stats.Errors++
			continue
		}
		stats.FinancialsStored++
	}

	stats.Elapsed = time.Since(start)
	return stats, nil
}

// FetchAndSaveTicker hämtar profil för ett enskilt bolag och sparar i DB.
// Används när bolaget inte redan finns i databasen.
func (s *SECStore) FetchAndSaveTicker(ctx context.Context, ticker string) (string, error) {
	// Hämta company list för att hitta CIK
	companies, err := s.sec.FetchCompanies(ctx)
	if err != nil {
		return "", fmt.Errorf("hämta bolag från SEC: %w", err)
	}

	var cik string
	for _, c := range companies {
		if strings.EqualFold(c.Ticker, ticker) {
			cik = c.ExternalID
			break
		}
	}
	if cik == "" {
		return "", fmt.Errorf("hittade inte %s i SEC", ticker)
	}

	// Hämta profil
	profile, err := s.sec.FetchProfile(ctx, cik)
	if err != nil {
		return "", fmt.Errorf("hämta profil för %s: %w", cik, err)
	}

	mc := &models.Company{
		ArgusID:  "C" + cik,
		Name:     profile.Description,
		Ticker:   ticker,
		CIK:      cik,
		Country:  profile.Country,
		Exchange: profile.Exchange,
	}

	_, err = s.db.UpsertCompany(ctx, mc)
	if err != nil {
		return "", fmt.Errorf("spara %s: %w", ticker, err)
	}

	return cik, nil
}

// FetchTickerProfile hämtar och sparar bolagsprofil för en CIK.
func (s *SECStore) FetchTickerProfile(ctx context.Context, cik, ticker string) error {
	profile, err := s.sec.FetchProfile(ctx, cik)
	if err != nil {
		return err
	}

	mc := &models.Company{
		ArgusID:  "C" + cik,
		Name:     profile.Description,
		Ticker:   ticker,
		CIK:      cik,
		Country:  profile.Country,
		Exchange: profile.Exchange,
	}

	_, err = s.db.UpsertCompany(ctx, mc)
	return err
}

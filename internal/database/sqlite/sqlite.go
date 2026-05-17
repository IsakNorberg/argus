package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/IsakNorberg/argus/pkg/models"
	_ "modernc.org/sqlite" // Pure Go SQLite, ingen CGO
)

// DB hanterar SQLite-databasen.
type DB struct {
	sql *sql.DB
}

// New skapar en ny databasanslutning.
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("öppna databas: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinga databas: %w", err)
	}

	return &DB{sql: db}, nil
}

// Init skapar tabeller om de inte finns.
func (db *DB) Init(ctx context.Context) error {
	schemaSQL := `
		CREATE TABLE IF NOT EXISTS companies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			argus_id TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			ticker TEXT DEFAULT '',
			cik TEXT DEFAULT '',
			org_number TEXT DEFAULT '',
			isin TEXT DEFAULT '',
			country TEXT DEFAULT '',
			industry TEXT DEFAULT '',
			sector TEXT DEFAULT '',
			exchange TEXT DEFAULT '',
			description TEXT DEFAULT '',
			website TEXT DEFAULT '',
			employees INTEGER DEFAULT 0,
			ceo TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS financials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			argus_id TEXT UNIQUE NOT NULL,
			company_id INTEGER NOT NULL,
			period TEXT NOT NULL,
			currency TEXT DEFAULT 'USD',
			revenue REAL DEFAULT 0,
			net_income REAL DEFAULT 0,
			total_assets REAL DEFAULT 0,
			total_debt REAL DEFAULT 0,
			equity REAL DEFAULT 0,
			eps REAL DEFAULT 0,
			ebitda REAL DEFAULT 0,
			free_cash_flow REAL DEFAULT 0,
			report_date DATETIME,
			filing_date DATETIME,
			source TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (company_id) REFERENCES companies(id),
			UNIQUE(company_id, period, source)
		);

		CREATE TABLE IF NOT EXISTS source_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			argus_id TEXT NOT NULL,
			source TEXT NOT NULL,
			last_fetched DATETIME DEFAULT CURRENT_TIMESTAMP,
			record_count INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			status TEXT DEFAULT 'ok',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(argus_id, source)
		);

		CREATE INDEX IF NOT EXISTS idx_company_ticker ON companies(ticker);
		CREATE INDEX IF NOT EXISTS idx_company_cik ON companies(cik);
		CREATE INDEX IF NOT EXISTS idx_company_orgnr ON companies(org_number);
		CREATE INDEX IF NOT EXISTS idx_company_isin ON companies(isin);
		CREATE INDEX IF NOT EXISTS idx_financials_company ON financials(company_id);
		CREATE INDEX IF NOT EXISTS idx_financials_period ON financials(period);
		CREATE INDEX IF NOT EXISTS idx_financials_source ON financials(source);

		CREATE TABLE IF NOT EXISTS price_quotes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			argus_id TEXT NOT NULL,
			ticker TEXT NOT NULL,
			price REAL DEFAULT 0,
			currency TEXT DEFAULT 'USD',
			market_cap INTEGER DEFAULT 0,
			volume INTEGER DEFAULT 0,
			change_percent REAL DEFAULT 0,
			source TEXT DEFAULT 'yahoo',
			quote_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(argus_id, source, quote_date)
		);

		CREATE INDEX IF NOT EXISTS idx_quote_ticker ON price_quotes(ticker);
		CREATE INDEX IF NOT EXISTS idx_quote_date ON price_quotes(quote_date);
	`

	_, err := db.sql.ExecContext(ctx, schemaSQL)
	if err != nil {
		return fmt.Errorf("skapa schema: %w", err)
	}

	log.Println("🗄️  Databas-schema skapat")
	return nil
}

// BeginTx startar en transaktion.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.sql.BeginTx(ctx, opts)
}

// CompanyByTicker slår upp ett bolag via ticker.
func (db *DB) CompanyByTicker(ctx context.Context, ticker string) (*models.Company, error) {
	c := &models.Company{}
	err := db.sql.QueryRowContext(ctx,
		`SELECT id, argus_id, name, ticker, cik, org_number, isin, country, industry, sector, exchange FROM companies WHERE ticker = ?`,
		ticker).Scan(&c.ID, &c.ArgusID, &c.Name, &c.Ticker, &c.CIK, &c.OrgNumber, &c.ISIN, &c.Country, &c.Industry, &c.Sector, &c.Exchange)
	if err != nil {
		return nil, fmt.Errorf("hitta %s: %w", ticker, err)
	}
	return c, nil
}

// CompanyByCIK slår upp ett bolag via CIK.
func (db *DB) CompanyByCIK(ctx context.Context, cik string) (*models.Company, error) {
	c := &models.Company{}
	err := db.sql.QueryRowContext(ctx,
		`SELECT id, argus_id, name, ticker, cik, org_number, isin, country, industry, sector, exchange FROM companies WHERE cik = ?`,
		cik).Scan(&c.ID, &c.ArgusID, &c.Name, &c.Ticker, &c.CIK, &c.OrgNumber, &c.ISIN, &c.Country, &c.Industry, &c.Sector, &c.Exchange)
	if err != nil {
		return nil, fmt.Errorf("hitta CIK %s: %w", cik, err)
	}
	return c, nil
}

// UpsertCompany upsertar ett bolag. Returnerar bolagets ID.
func (db *DB) UpsertCompany(ctx context.Context, c *models.Company) (int64, error) {
	query := `
		INSERT INTO companies (argus_id, name, ticker, cik, org_number, isin,
			country, industry, sector, exchange, description, website, employees, ceo)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(argus_id) DO UPDATE SET
			name = excluded.name,
			ticker = excluded.ticker,
			cik = excluded.cik,
			org_number = excluded.org_number,
			isin = excluded.isin,
			country = excluded.country,
			industry = excluded.industry,
			sector = excluded.sector,
			exchange = excluded.exchange,
			description = excluded.description,
			website = excluded.website,
			employees = excluded.employees,
			ceo = excluded.ceo,
			updated_at = CURRENT_TIMESTAMP
	`

	var id int64
	err := db.sql.QueryRowContext(ctx, `SELECT id FROM companies WHERE argus_id = ?`, c.ArgusID).Scan(&id)

	if err == sql.ErrNoRows {
		result, err := db.sql.ExecContext(ctx, query,
			c.ArgusID, c.Name, c.Ticker, c.CIK, c.OrgNumber, c.ISIN,
			c.Country, c.Industry, c.Sector, c.Exchange, c.Description,
			c.Website, c.Employees, c.CEO,
		)
		if err != nil {
			return 0, fmt.Errorf("insert company: %w", err)
		}
		return result.LastInsertId()
	} else if err != nil {
		return 0, fmt.Errorf("sök company: %w", err)
	}

	_, err = db.sql.ExecContext(ctx, query,
		c.ArgusID, c.Name, c.Ticker, c.CIK, c.OrgNumber, c.ISIN,
		c.Country, c.Industry, c.Sector, c.Exchange, c.Description,
		c.Website, c.Employees, c.CEO,
	)
	if err != nil {
		return 0, fmt.Errorf("update company: %w", err)
	}

	return id, nil
}

// UpsertFinancials upsertar finansiell data.
func (db *DB) UpsertFinancials(ctx context.Context, f *models.Financials) error {
	query := `
		INSERT INTO financials (argus_id, company_id, period, currency, revenue,
			net_income, total_assets, total_debt, equity, eps, ebitda,
			free_cash_flow, report_date, filing_date, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(company_id, period, source) DO UPDATE SET
			revenue = excluded.revenue,
			net_income = excluded.net_income,
			total_assets = excluded.total_assets,
			total_debt = excluded.total_debt,
			equity = excluded.equity,
			eps = excluded.eps,
			ebitda = excluded.ebitda,
			free_cash_flow = excluded.free_cash_flow,
			report_date = excluded.report_date,
			filing_date = excluded.filing_date,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.sql.ExecContext(ctx, query,
		f.ArgusID, f.CompanyID, f.Period, f.Currency, f.Revenue,
		f.NetIncome, f.TotalAssets, f.TotalDebt, f.Equity, f.EPS, f.EBITDA,
		f.FreeCashFlow, f.ReportDate, f.FilingDate, f.Source,
	)
	if err != nil {
		return fmt.Errorf("upsert financials: %w", err)
	}
	return nil
}

// UpsertPriceQuote upsertar en dagprisnotering.
func (db *DB) UpsertPriceQuote(ctx context.Context, q *models.PriceQuote) error {
	query := `
		INSERT INTO price_quotes (argus_id, ticker, price, currency, market_cap,
			volume, change_percent, source, quote_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(argus_id, source) DO UPDATE SET
			price = excluded.price,
			volume = excluded.volume,
			change_percent = excluded.change_percent,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.sql.ExecContext(ctx, query,
		q.ArgusID, q.Ticker, q.Price, q.Currency,
		q.MarketCap, q.Volume, q.ChangePercent,
		q.Source, q.QuoteDate,
	)
	if err != nil {
		return fmt.Errorf("upsert price quote: %w", err)
	}
	return nil
}

// Close stänger databasen.
func (db *DB) Close() error {
	return db.sql.Close()
}

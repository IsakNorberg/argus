package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/IsakNorberg/argus/internal/database/sqlite"
	"github.com/IsakNorberg/argus/pkg/models"
)

// DB är huvudinterfacet för databasoperationer.
// Implementeras av SQLite (built-in) och PostgreSQL (framtida).
type DB interface {
	Init(ctx context.Context) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	CompanyByTicker(ctx context.Context, ticker string) (*models.Company, error)
	CompanyByCIK(ctx context.Context, cik string) (*models.Company, error)
	UpsertCompany(ctx context.Context, c *models.Company) (int64, error)
	UpsertFinancials(ctx context.Context, f *models.Financials) error
	UpsertPriceQuote(ctx context.Context, q *models.PriceQuote) error
	Close() error
}

var currentDialect = "sqlite"

// Dialect returnerar den aktiva SQL-dialekten.
func Dialect() string {
	return currentDialect
}

// New skapar en databas baserat på miljövariabler.
// ARGUS_DB_DRIVER: "sqlite" (default) eller "postgres"
// ARGUS_DB_URL: anslutningssträng (ignoreras för sqlite)
// dbPath: SQLite-filens sökväg (ignoreras för postgres)
func New(dbPath string) (DB, error) {
	driver := os.Getenv("ARGUS_DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	switch driver {
	case "postgres", "pg":
		url := os.Getenv("ARGUS_DB_URL")
		if url == "" {
			url = "postgres://argus:argus@localhost:5432/argus?sslmode=disable"
		}
		currentDialect = "postgres"
		// TODO: implementera postgres.New(url)
		return nil, fmt.Errorf("postgres-driver planerad men inte implementerad ännu")
	default:
		currentDialect = "sqlite"
		return sqlite.New(dbPath)
	}
}

// Placeholders genererar SQL-placeholders för en given dialekt.
// SQLite: ?, PostgreSQL: $1, $2, ...
func Placeholders(dialect string, count int) string {
	if dialect == "postgres" {
		var b strings.Builder
		for i := 0; i < count; i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, "$%d", i+1)
		}
		return b.String()
	}
	// SQLite
	return strings.Repeat("?,", count-1) + "?"
}

// UpsertClause returnerar rätt ON CONFLICT-syntax för dialekten.
// Båda stödjer ON CONFLICT — men PostgreSQL kräver parentes.
func UpsertClause(dialect string, conflictCols string, updateSets string) string {
	if dialect == "postgres" {
		return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", conflictCols, updateSets)
	}
	return fmt.Sprintf("ON CONFLICT(%s) DO UPDATE SET %s", conflictCols, updateSets)
}

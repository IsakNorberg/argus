package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/IsakNorberg/argus/pkg/models"
)

// DB hanterar databasen.
type DB struct {
	path string
}

// New skapar en ny databasanslutning.
func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}
	return &DB{path: dbPath}, nil
}

// Init skapar tabeller om de inte finns.
func (db *DB) Init() error {
	// TODO: SQLite eller PostgreSQL setup
	// CREATE TABLE IF NOT EXISTS companies ...
	// CREATE TABLE IF NOT EXISTS financials ...
	return fmt.Errorf("inte implementerad ännen")
}

// UpsertCompany upsertar ett bolag.
func (db *DB) UpsertCompany(c *models.Company) error {
	return fmt.Errorf("inte implementerad ännen")
}

// UpsertFinancials upsertar finansiell data.
func (db *DB) UpsertFinancials(f *models.Financials) error {
	return fmt.Errorf("inte implementerad ännen")
}

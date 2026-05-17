package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources"
	"github.com/IsakNorberg/argus/internal/sources/sec"
	"github.com/IsakNorberg/argus/internal/sources/yahoo"
	"github.com/IsakNorberg/argus/internal/store"
	"github.com/IsakNorberg/argus/pkg/models"

	// Nordic sources
	"github.com/IsakNorberg/argus/internal/sources/ares"
	"github.com/IsakNorberg/argus/internal/sources/bolagsverket"
	"github.com/IsakNorberg/argus/internal/sources/brreg"
	"github.com/IsakNorberg/argus/internal/sources/companieshouse"
	"github.com/IsakNorberg/argus/internal/sources/cro"
	"github.com/IsakNorberg/argus/internal/sources/cvr"
	"github.com/IsakNorberg/argus/internal/sources/edinet"
	"github.com/IsakNorberg/argus/internal/sources/inpi"
	"github.com/IsakNorberg/argus/internal/sources/krs"
	"github.com/IsakNorberg/argus/internal/sources/kvk"
	"github.com/IsakNorberg/argus/internal/sources/prh"
	"github.com/IsakNorberg/argus/internal/sources/sedar"
	"github.com/IsakNorberg/argus/internal/sources/zefix"
)

var version = "0.4.3"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "fetch", "f":
		runFetch(os.Args[2:])
	case "status", "s":
		runStatus()
	case "version", "-v", "--version":
		fmt.Printf("Argus %s\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`
🏰 Argus v0.4.3 — The Watcher

USAGE
  argus fetch <source> [options]    Fetch financial data
  argus status                      Show all source statuses  
  argus version                     Show version

SOURCES
  🇺🇸 sec          SEC EDGAR (US)
  🇫🇮 prh          PRH (Finland)
  🇳🇴 brreg        Brreg (Norway)
  🇩🇰 cvr          CVR (Denmark)
  🇸🇪 bolagsverket Bolagsverket (Sweden)
  🇬🇧 ch           Companies House (UK)
  🇨🇦 sedar        SEDAR (Canada)
  🇯🇵 edinet        EDINET (Japan)
  🇩🇪 krs          KRS/EMS (Germany)
  🇫🇷 inpi         INPI (France)
  🇳🇱 kvk          KVK (Netherlands)
  🇨🇭 zefix        Zefix (Switzerland)
  🇮🇪 cro          CRO (Ireland)
  🇦🇹 ares         ARES (Austria)
  🇵🇱 krs          KRS (Poland)
  🌐 yahoo        Yahoo Finance (prices)

EXAMPLES
  argus fetch sec --all
  argus fetch sec --ticker AAPL,MSFT
  argus fetch prh --all
  argus fetch brreg --ticker NOK
  argus status
`)
}

func runStatus() {
	type src struct {
		flag    string
		name    string
		impl    string
		status  string
		ticker  string
	}

	registry := []src{
		{"🇺🇸", "SEC EDGAR", "sec", "✅", "US"},
		{"🇫🇮", "PRH", "prh", "✅", "FI"},
		{"🇳🇴", "Brreg", "brreg", "✅", "NO"},
		{"🇩🇰", "CVR", "cvr", "🟡", "DK"},
		{"🇸🇪", "Bolagsverket", "bolagsverket", "⏳", "SE"},
		{"🇬🇧", "Companies House", "companieshouse", "⏳", "GB"},
		{"🇨🇦", "SEDAR+", "sedar", "⏳", "CA"},
		{"🇯🇵", "EDINET", "edinet", "⏳", "JP"},
		{"🇩🇪", "KRS", "krs_de", "⏳", "DE"},
		{"🇫🇷", "INPI", "inpi", "⏳", "FR"},
		{"🇳🇱", "KVK", "kvk", "⏳", "NL"},
		{"🇨🇭", "Zefix", "zefix", "⏳", "CH"},
		{"🇮🇪", "CRO", "cro", "⏳", "IE"},
		{"🇦🇹", "ARES", "ares", "⏳", "AT"},
		{"🇵🇱", "KRS", "krs", "⏳", "PL"},
		{"🌐", "Yahoo", "yahoo", "🟡", ""},
	}

	fmt.Println("🏛️  Argus — Register Sources")
	fmt.Println()

	// Table header
	fmt.Printf("  %-4s %-20s %-10s %-8s %s\n", "", "Name", "ID", "Status", "Exchange")
	fmt.Println("  " + strings.Repeat("─", 60))

	for _, s := range registry {
		fmt.Printf("  %s %-20s %-10s %-8s %s\n",
			s.flag, s.name, s.impl, s.status, s.ticker)
	}

	fmt.Println()
	fmt.Println("✅ = Implemented   🟡 = Partial   ⏳ = Stub")
	fmt.Println()
	fmt.Printf("💾 Database: %s\n", dbPath())
	fmt.Printf("🔧 Dialect:  %s\n", database.Dialect())
}

func runFetch(args []string) {
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
	tickers := fetchCmd.String("ticker", "", "Comma-separated tickers")
	all := fetchCmd.Bool("all", false, "Fetch ALL companies (slow)")

	fetchCmd.Parse(args)
	if fetchCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: argus fetch <source> [--ticker X | --all]")
		os.Exit(1)
	}

	sourceName := fetchCmd.Arg(0)

	// Validate source
	if !*all && *tickers == "" {
		fmt.Fprintln(os.Stderr, "Provide --ticker or --all")
		os.Exit(1)
	}

	db := setupDB()
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// SEC special case
	if sourceName == "sec" {
		secClient := sec.New()
		s := store.New(db, secClient)
		runSecFetch(ctx, s, *tickers, *all)
		return
	}

	if sourceName == "yahoo" {
		if !*all && *tickers == "" {
			fmt.Fprintln(os.Stderr, "Yahoo requires --ticker")
			os.Exit(1)
		}
		yahooClient := yahoo.New()
		yahooStore := store.NewYahooStore(db, yahooClient)
		runYahooFetch(ctx, yahooStore, *tickers)
		return
	}

	// All other sources via registry
	src := resolveSource(sourceName)
	if src == nil {
		fmt.Fprintf(os.Stderr, "Unknown source: %s\n", sourceName)
		os.Exit(1)
	}

	genericFetch(ctx, db, src, *tickers, *all)
}

func resolveSource(name string) sources.Source {
	m := map[string]func() sources.Source{
		"prh":            func() sources.Source { return prh.New() },
		"brreg":          func() sources.Source { return brreg.New() },
		"cvr":            func() sources.Source { return cvr.New() },
		"bolagsverket":   func() sources.Source { return bolagsverket.New() },
		"companieshouse": func() sources.Source { return companieshouse.New() },
		"sedar":          func() sources.Source { return sedar.New() },
		"edinet":         func() sources.Source { return edinet.New() },
		"krs":            func() sources.Source { return krs.New() },
		"inpi":           func() sources.Source { return inpi.New() },
		"kvk":            func() sources.Source { return kvk.New() },
		"zefix":          func() sources.Source { return zefix.New() },
		"cro":            func() sources.Source { return cro.New() },
		"ares":           func() sources.Source { return ares.New() },
	}
	f, ok := m[name]
	if !ok {
		return nil
	}
	return f()
}

func genericFetch(ctx context.Context, db database.DB, src sources.Source, tickers string, all bool) {
	fmt.Printf("⏳ Fetching from %s (%s)...\n", src.Name(), src.Name())

	companies, err := src.FetchCompanies(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %d companies found\n", len(companies))

	if all && len(companies) > 200 {
		fmt.Printf("⚠️  Limiting to 200 of %d (use --ticker for specific)\n", len(companies))
		companies = companies[:200]
	}

	// Filter by tickers if specified
	if tickers != "" {
		tickerSet := make(map[string]bool)
		for _, t := range splitTickers(tickers) {
			tickerSet[t] = true
		}
		filtered := companies[:0]
		for _, c := range companies {
			if tickerSet[strings.ToUpper(c.Ticker)] {
				filtered = append(filtered, c)
			}
		}
		companies = filtered
		fmt.Printf("   Filtered to %d matching companies\n", len(companies))
	}

	// Save companies
	saved := 0
	for _, c := range companies {
		if _, err := db.UpsertCompany(ctx, &models.Company{
			Name:       c.Name,
			Ticker:     c.Ticker,
			Country:    c.Country,
			Exchange:   c.Exchange,
			Industry:   c.Industry,
		}); err != nil {
			continue
		}
		saved++
	}

	fmt.Printf("✅ %d companies saved to DB\n", saved)

	// Fetch financials for saved companies
	if all {
		fmt.Println("\n⏳ Fetching financials...")
		fetched := 0
		errors := 0
		for _, c := range companies {
			if c.ExternalID == "" {
				continue
			}
			financials, err := src.FetchFinancials(ctx, c.ExternalID)
			if err != nil {
				errors++
				continue
			}
		for _, f := range financials {
			if err := db.UpsertFinancials(ctx, &models.Financials{
				CompanyID:  0, // TODO: lookup by ExternalID
				Period:     f.Period,
				Currency:   f.Currency,
				Revenue:    f.Revenue,
				NetIncome:  f.NetIncome,
				TotalAssets: f.TotalAssets,
				Source:     f.Source,
			}); err != nil {
				continue
			}
			fetched++
		}
		}
		fmt.Printf("✅ %d financial periods saved (%d errors)\n\n", fetched, errors)
	}

	// Print summary table
	fmt.Println("\n📊 Summary:")
	fmt.Printf("  Source:     %s\n", src.Name())
	fmt.Printf("  Companies:  %d fetched, %d saved\n", len(companies), saved)
}

func runSecFetch(ctx context.Context, s *store.SECStore, tickers string, all bool) {
	if all {
		fmt.Println("⏳ Fetching ALL companies from SEC EDGAR...")
		stats, tickerToCIK, err := s.IngestCompanies(ctx)
		if err != nil {
			fmt.Printf("❌ IngestCompanies: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Companies: %d fetched, %d saved, %d errors (%v)\n",
			stats.CompaniesFetched, stats.CompaniesStored, stats.Errors, stats.Elapsed)

		fmt.Println("\n Examples of saved companies (first 5):")
		count := 0
		for t := range tickerToCIK {
			if count >= 5 {
				break
			}
			fmt.Printf("  %s → CIK %s\n", t, tickerToCIK[t])
			count++
		}
		fmt.Printf("  ...and %d more\n", len(tickerToCIK)-5)
		return
	}

	tickerList := splitTickers(tickers)

	fmt.Printf("⏳ Fetching financial data for %d tickers...\n", len(tickerList))
		for i, t := range tickerList {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(tickerList), t)

		cik, err := s.GetCIKByTicker(ctx, t)
		if err != nil {
			fmt.Printf("  ⚠️  Not in DB, fetching from SEC...\n")
			cik, err = s.FetchAndSaveTicker(ctx, t)
			if err != nil {
				fmt.Printf("  ❌ CIK lookup: %v\n", err)
				continue
			}
			fmt.Printf("  ✅ Profile fetched, CIK %s\n", cik)
		} else {
			fmt.Printf("  ✅ CIK %s (from DB)\n", cik)
		}

		finStats, err := s.FetchAndSaveFinancials(ctx, cik)
		if err != nil {
			fmt.Printf("  ❌ Financials: %v\n", err)
			continue
		}
		fmt.Printf("  ✅ %d periods saved (%v)\n", finStats.FinancialsStored, finStats.Elapsed)
	}

	fmt.Println("\n✅ Done!")
}

func runYahooFetch(ctx context.Context, s *store.YahooStore, tickers string) {
	tickerList := splitTickers(tickers)

	fmt.Printf("⏳ Fetching prices for %d companies from Yahoo Finance...\n", len(tickerList))
	stats, err := s.FetchAndSaveQuotes(ctx, tickerList)
	if err != nil {
		fmt.Printf("❌ Yahoo: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ %d tickers, %d saved, %d errors (%v)\n",
		stats.Fetched + stats.Errors, stats.Stored, stats.Errors, stats.Elapsed)
}

func splitTickers(raw string) []string {
	var result []string
	for _, t := range strings.Split(raw, ",") {
		t = strings.TrimSpace(strings.ToUpper(t))
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func dbPath() string {
	p := os.Getenv("ARGUS_DB_PATH")
	if p == "" {
		p = "./data/argus.db"
	}
	return p
}

func setupDB() database.DB {
	dbPath := dbPath()
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("❌ Database error: %v", err)
	}

	ctx := context.Background()
	if err := db.Init(ctx); err != nil {
		log.Fatalf("❌ Schema error: %v", err)
	}

	return db
}

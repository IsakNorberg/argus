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
	"github.com/IsakNorberg/argus/internal/sources/prh"
	"github.com/IsakNorberg/argus/internal/sources/brreg"
	"github.com/IsakNorberg/argus/internal/sources/cvr"
)

var version = "0.4.2"

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetch":
		runFetch()
	case "status":
		runStatus()
	case "--version", "-v", "version":
		fmt.Printf("Argus %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "❌ Okänt kommando: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`🏛️  Argus v%s — Den med 100 ögon

Användning:
  argus <kommando> [flaggor]

Kommandon:
  fetch    Hämtar finansiell data från en källa
  status   Visa status för alla registerkällor
  version  Visa version

Exempel:
  argus fetch sec --ticker AAPL
  argus fetch sec --all
  argus fetch sec --ticker AAPL,MSFT,GOOGL
  argus status
`, version)
}

// --- FETCH KOMMANDO ---

func runFetch() {
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
	tickers := fetchCmd.String("ticker", "", "Ticker(s) att hämta, kommar-separerad")
	all := fetchCmd.Bool("all", false, "Hämta ALLA bolag från källan")

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "❌ Ange en källa: sec, yahoo")
		fmt.Fprintln(os.Stderr, "   Använd: argus fetch <källa> --ticker AAPL")
		os.Exit(1)
	}

	sourceName := os.Args[2]
	fetchCmd.Parse(os.Args[3:])

	if !*all && *tickers == "" {
		fmt.Fprintln(os.Stderr, "❌ Ange --ticker <TICKER> eller --all")
		os.Exit(1)
	}

	db := setupDB()
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	switch sourceName {
	case "sec":
		secClient := sec.New()
		s := store.New(db, secClient)
		runSecFetch(ctx, s, *tickers, *all)
	case "yahoo":
		yahooClient := yahoo.New()
		yahooStore := store.NewYahooStore(db, yahooClient)
		runYahooFetch(ctx, yahooStore, *tickers)
	case "prh":
		prhClient := prh.New()
		runNordicFetch(ctx, db, prhClient, *tickers, *all, "prh", "FI", "HEX")
	case "brreg":
		brregClient := brreg.New()
		runNordicFetch(ctx, db, brregClient, *tickers, *all, "brreg", "NO", "OSE")
	case "cvr":
		cvrClient := cvr.New()
		runNordicFetch(ctx, db, cvrClient, *tickers, *all, "cvr", "DK", "CSE")
	default:
		fmt.Fprintf(os.Stderr, "❌ Okänd källa: %s\n", sourceName)
		fmt.Fprintln(os.Stderr, "   tillgängliga: sec, yahoo, prh, brreg, cvr")
		os.Exit(1)
	}
}

func runYahooFetch(ctx context.Context, s *store.YahooStore, tickers string) {
	tickerList := splitTickers(tickers)

	fmt.Printf("⏳ Hämtar priser för %d bolag från Yahoo Finance ...\n", len(tickerList))
	stats, err := s.FetchAndSaveQuotes(ctx, tickerList)
	if err != nil {
		fmt.Printf("❌ Yahoo: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ %d tickers, %d sparade, %d fel (%v)\n",
		stats.Fetched + stats.Errors, stats.Stored, stats.Errors, stats.Elapsed)
}

func runSecFetch(ctx context.Context, s *store.SECStore, tickers string, all bool) {
	if all {
		fmt.Println("⏳ Hämtar ALLA bolag från SEC EDGAR...")
		stats, tickerToCIK, err := s.IngestCompanies(ctx)
		if err != nil {
			fmt.Printf("❌ IngestCompanies: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Bolag: %d hämtade, %d sparade, %d fel (%v)\n",
			stats.CompaniesFetched, stats.CompaniesStored, stats.Errors, stats.Elapsed)

		// Visa några exempel
		fmt.Println("\nExempel på sparade bolag (första 5):")
		count := 0
		for t := range tickerToCIK {
			if count >= 5 {
				break
			}
			fmt.Printf("  %s → CIK %s\n", t, tickerToCIK[t])
			count++
		}
		fmt.Printf("  ...och %d till\n", len(tickerToCIK)-5)
		return
	}

	// Hämta specifika tickers med auto-fallback
	tickerList := splitTickers(tickers)

	fmt.Printf("⏳ Hämtar finansiell data för %d bolag...\n", len(tickerList))
	for i, t := range tickerList {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(tickerList), t)

		cik, err := s.GetCIKByTicker(ctx, t)
		if err != nil {
			// Fallback: hämta profil från SEC och spara i DB
			fmt.Printf("  ⚠️  Ej i DB, hämtar profil från SEC...\n")
			cik, err = s.FetchAndSaveTicker(ctx, t)
			if err != nil {
				fmt.Printf("  ❌ CIK lookup: %v\n", err)
				continue
			}
			fmt.Printf("  ✅ Profile hämtad, CIK %s\n", cik)
		} else {
			fmt.Printf("  ✅ CIK %s (från DB)\n", cik)
		}

		finStats, err := s.FetchAndSaveFinancials(ctx, cik)
		if err != nil {
			fmt.Printf("  ❌ Financials: %v\n", err)
			continue
		}
		fmt.Printf("  ✅ %d perioder sparade (%v)\n", finStats.FinancialsStored, finStats.Elapsed)
	}

	fmt.Println("\n✅ Klart!")
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

// --- STATUS KOMMANDO ---

func runNordicFetch(ctx context.Context, db database.DB, src sources.Source, tickers string, all bool, srcName, country, exchange string) {
	fmt.Printf("⏳ Hämtar bolag från %s...\n", strings.ToUpper(srcName))

	// Enkel wrapper: hämta bolag → spara → hämta finansiell data
	companies, err := src.FetchCompanies(ctx)
	if err != nil {
		fmt.Printf("❌ Hämta bolag: %v\n", err)
		os.Exit(1)
	}

	if !all && tickers != "" {
		// Filtrera efter angivna tickers
		tickerList := splitTickers(tickers)
		tickerSet := make(map[string]bool)
		for _, t := range tickerList {
			tickerSet[t] = true
		}
		var filtered []sources.Company
		for _, c := range companies {
			if tickerSet[strings.ToUpper(c.Ticker)] {
				filtered = append(filtered, c)
			}
		}
		companies = filtered
	}

	fmt.Printf("✅ %d bolag hittade\n", len(companies))

	if all && len(companies) > 100 {
		fmt.Printf("⚠️  Sparar första 100 av %d (använd --ticker för specifika)\n", len(companies))
		companies = companies[:100]
	}

	// Spara bolag
	saved := 0
	for _, c := range companies {
		c.Country = country
		c.Exchange = exchange

		// Upsert company
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

	fmt.Printf("✅ %d bolag sparade till DB\n", saved)
}

func runStatus() {
	fmt.Println("🏛️  Argus — Registerkällor")
	fmt.Println()
	sources := []struct {
		flag   string
		name   string
		status string
	}{
		{"🇺🇸", "SEC EDGAR", "🟢 Implementerad"},
		{"🇫🇮", "PRH", "🟢 Implementerad"},
		{"🇳🇴", "Brreg", "🟢 Implementerad"},
		{"🇩🇰", "CVR", "🟡 Bolag OK, finansiell TBD"},
		{"🇸🇪", "Bolagsverket", "🟡 Stub (API-nyckel)"},
		{"🇬🇧", "Companies House", "⏳ Stub"},
		{"🇨🇦", "SEDAR+", "⏳ Stub"},
		{"🇯🇵", "EDINET", "⏳ Stub"},
		{"🇫🇷", "INPI", "⏳ Stub"},
		{"🇩🇪", "KRS", "⏳ Stub"},
		{"🇳🇱", "KVK", "⏳ Stub"},
		{"🇨🇭", "Zefix", "⏳ Stub"},
		{"🇮🇪", "CRO", "⏳ Stub"},
		{"🇦🇹", "ARES", "⏳ Stub"},
		{"🇵🇱", "KRS", "⏳ Stub"},
	}

	for _, s := range sources {
		fmt.Printf("  %s %-20s — %s\n", s.flag, s.name, s.status)
	}

	fmt.Println()
	fmt.Println("📈 Yahoo — dagliga kurser (tillgänglig)")
	fmt.Println()
	fmt.Printf("💾 Databas: %s\n", dbPath())
	fmt.Printf("🔧 Dialect: %s\n", database.Dialect())
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
		log.Fatalf("❌ Databasfel: %v", err)
	}

	ctx := context.Background()
	if err := db.Init(ctx); err != nil {
		log.Fatalf("❌ Schema-fel: %v", err)
	}

	return db
}

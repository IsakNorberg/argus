package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources/sec"
	"github.com/IsakNorberg/argus/internal/store"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := database.New(":memory:")
	if err != nil {
		fmt.Println("❌ DB:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Verifiera att det är rätt implementation
	fmt.Printf("📋 Dialect: %s\n", database.Dialect())

	if err := db.Init(ctx); err != nil {
		fmt.Println("❌ Init:", err)
		os.Exit(1)
	}

	secClient := sec.New()
	s := store.New(db, secClient)

	fmt.Println("\n=== Testa IngestCompanies ===")
	stats, _, err := s.IngestCompanies(ctx)
	if err != nil {
		fmt.Println("❌ IngestCompanies:", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Bolag: %d hämtade, %d sparade, %d fel (%v)\n",
		stats.CompaniesFetched, stats.CompaniesStored, stats.Errors, stats.Elapsed)

	fmt.Println("\n=== Testa FetchAndSaveFinancials (Apple via ticker) ===")
	cik, err := s.GetCIKByTicker(ctx, "AAPL")
	if err != nil {
		fmt.Println("❌ GetCIKByTicker:", err)
		os.Exit(1)
	}
	fmt.Printf("✅ AAPL → CIK %s\n", cik)

	finStats, err := s.FetchAndSaveFinancials(ctx, cik)
	if err != nil {
		fmt.Println("❌ Financials Apple:", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Apple: %d perioder hämtade, %d sparade, %d fel (%v)\n",
		finStats.FinancialsFetched, finStats.FinancialsStored, finStats.Errors, finStats.Elapsed)

	fmt.Println("\n🎉 Alla test passade!")
}

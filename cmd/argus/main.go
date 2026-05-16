package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources"
	"github.com/IsakNorberg/argus/internal/sources/allabolag"
	"github.com/IsakNorberg/argus/internal/sources/bolagsverket"
	"github.com/IsakNorberg/argus/internal/sources/eodhd"
	"github.com/IsakNorberg/argus/internal/sources/fi"
	"github.com/IsakNorberg/argus/internal/sources/morningstar"
	"github.com/IsakNorberg/argus/internal/sources/sec"
	"github.com/IsakNorberg/argus/internal/sources/yahoo"
)

var version = "0.1.0"

func main() {
	log.Println("🏛️  Argus — Den med 100 ögon")
	log.Printf("Version %s\n", version)

	ctx := context.Background()

	// Initiera databas
	dbPath := "./data/argus.db"
	os.MkdirAll("./data", 0755)

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatal("❌ Databasfel:", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		log.Fatal("❌ Schema-fel:", err)
	}

	// Registrera alla källor
	sources_ := []sources.Source{
		sec.New(os.Getenv("ARGUS_EMAIL")),       // SEC (ingen nyckel)
		eodhd.New(os.Getenv("EODHD_API_KEY")),    // EODHD
		bolagsverket.New(os.Getenv("BV_API_KEY")), // Bolagsverket
		yahoo.New(),                              // Yahoo Finance
		allabolag.New(),                          // Allabolag.se
		morningstar.New(),                        // Morningstar
		fi.New(),                                 // Finansinspeksen
	}

	fmt.Printf("\n✅ Argus är redo!\n\n")
	fmt.Printf("Databas: %s\n", dbPath)
	fmt.Printf("Källor: %d registrerade\n", len(sources_))
	fmt.Println("  📡 SEC (EDGAR)")
	fmt.Println("  📡 EODHD")
	fmt.Println("  📡 Bolagsverket")
	fmt.Println("  📡 Yahoo Finance")
	fmt.Println("  📡 Allabolag.se")
	fmt.Println("  📡 Morningstar")
	fmt.Println("  📡 Finansinspeksen")
	fmt.Println("\nNästa steg: Implementera fetchers per källa.")
}

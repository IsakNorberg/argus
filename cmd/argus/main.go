package main

import (
	"context"
	"fmt"
	"log"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources"
	"github.com/IsakNorberg/argus/internal/sources/brreg"
	"github.com/IsakNorberg/argus/internal/sources/bolagsverket"
	"github.com/IsakNorberg/argus/internal/sources/companieshouse"
	"github.com/IsakNorberg/argus/internal/sources/cvr"
	"github.com/IsakNorberg/argus/internal/sources/inpi"
	"github.com/IsakNorberg/argus/internal/sources/kvk"
	"github.com/IsakNorberg/argus/internal/sources/prh"
	"github.com/IsakNorberg/argus/internal/sources/sec"
	"github.com/IsakNorberg/argus/internal/sources/yahoo"
)

var version = "0.2.0"

func main() {
	log.Println("🏛️  Argus — Den med 100 ögon")
	log.Printf("Version %s\n", version)

	ctx := context.Background()

	// Initiera databas
	dbPath := "./data/argus.db"

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatal("❌ Databasfel: ", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		log.Fatal("❌ Schema-fel: ", err)
	}

	// Registrera alla källor med öppen API
	registerSources := []sources.Source{
		sec.New(),              // 🇺🇸 SEC (EDGAR) — helt gratis
		bolagsverket.New(),     // 🇸🇪 Bolagsverket — nyckel krävs
		brreg.New(),            // 🇳🇴 Brreg — grundläggande auth
		cvr.New(),              // 🇩🇰 CVR — helt gratis
		prh.New(),              // 🇫🇮 PRH — helt gratis
		companieshouse.New(),   // 🇬🇧 Companies House — helt gratis
		inpi.New(),             // 🇫🇷 INPI — gratis
		kvk.New(),              // 🇳🇱 KVK — nyckel krävs
	}

	// Yahoo för dagens kurs (separat)
	_ = yahoo.New() // 📈 Yahoo — KURSER

	fmt.Printf("\n✅ Argus är redo!\n\n")
	fmt.Printf("Databas: %s\n", dbPath)
	fmt.Printf("Register: %d källor med API\n\n", len(registerSources))
	fmt.Println("  🟢 🇺🇸 SEC (EDGAR) — helt gratis")
	fmt.Println("  🟡 🇸🇪 Bolagsverket — API-nyckel")
	fmt.Println("  🟡 🇳🇴 Brreg — grundläggande auth")
	fmt.Println("  🟢 🇩🇰 CVR — helt gratis")
	fmt.Println("  🟢 🇫🇮 PRH — helt gratis")
	fmt.Println("  🟢 🇬🇧 Companies House — helt gratis")
	fmt.Println("  🟡 🇫🇷 INPI — gratis")
	fmt.Println("  🟡 🇳🇱 KVK — API-nyckel")
	fmt.Println("\n  📈 Yahoo — dagens kurs")
	fmt.Println("\nNästa steg: Implementera fetchers.")
}

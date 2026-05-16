package main

import (
	"context"
	"fmt"
	"log"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources"
	"github.com/IsakNorberg/argus/internal/sources/brreg"
	"github.com/IsakNorberg/argus/internal/sources/bundesanzeiger"
	"github.com/IsakNorberg/argus/internal/sources/bolagsverket"
	"github.com/IsakNorberg/argus/internal/sources/companieshouse"
	"github.com/IsakNorberg/argus/internal/sources/cvr"
	"github.com/IsakNorberg/argus/internal/sources/esef"
	"github.com/IsakNorberg/argus/internal/sources/inpi"
	"github.com/IsakNorberg/argus/internal/sources/kvk"
	"github.com/IsakNorberg/argus/internal/sources/prh"
	"github.com/IsakNorberg/argus/internal/sources/registro_mercantil"
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

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatal("❌ Databasfel: ", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		log.Fatal("❌ Schema-fel: ", err)
	}

	// Registrera alla officiella register (fundamentals)
	officialSources := []sources.Source{
		sec.New(),              // 🇺🇸 SEC (EDGAR)
		bolagsverket.New(),     // 🇸🇪 Bolagsverket
		brreg.New(),            // 🇳🇴 Brønnøysundregistrene
		cvr.New(),              // 🇩🇰 CVR / Erhvervsstyrelsen
		prh.New(),              // 🇫🇮 PRH
		companieshouse.New(),   // 🇬🇧 Companies House
		bundesanzeiger.New(),   // 🇩🇪 Bundesanzeiger
		inpi.New(),             // 🇫🇷 INPI / BODACC
		kvk.New(),              // 🇳🇱 KVK
		esef.New(),             // 🇪🇺 ESEF (EU XBRL)
		registro_mercantil.New(), // 🇪🇸 Registro Mercantil
	}

	// Yahoo för dagens kurs (separat)
	_ = yahoo.New() // 📈 Yahoo Finance — KURSER

	fmt.Printf("\n✅ Argus är redo!\n\n")
	fmt.Printf("Databas: %s\n", dbPath)
	fmt.Printf("Register: %d officiella källor (fundamentals)\n\n", len(officialSources))
	fmt.Println("  🇺🇸 SEC (EDGAR)")
	fmt.Println("  🇸🇪 Bolagsverket")
	fmt.Println("  🇳🇴 Brønnøysundregistrene")
	fmt.Println("  🇩🇰 CVR / Erhvervsstyrelsen")
	fmt.Println("  🇫🇮 PRH")
	fmt.Println("  🇬🇧 Companies House")
	fmt.Println("  🇩🇪 Bundesanzeiger")
	fmt.Println("  🇫🇷 INPI / BODACC")
	fmt.Println("  🇳🇱 KVK")
	fmt.Println("  🇪🇺 ESEF (EU XBRL)")
	fmt.Println("  🇪🇸 Registro Mercantil")
	fmt.Println("\n  📈 Yahoo Finance — dagens kurs (ej fundamentals)")
	fmt.Println("\nNästa steg: Implementera fetchers per register.")
}

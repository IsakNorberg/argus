package main

import (
	"context"
	"fmt"
	"log"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources"
	"github.com/IsakNorberg/argus/internal/sources/bolagsverket"
	"github.com/IsakNorberg/argus/internal/sources/brreg"
	"github.com/IsakNorberg/argus/internal/sources/edinet"
	"github.com/IsakNorberg/argus/internal/sources/sec"
	"github.com/IsakNorberg/argus/internal/sources/yahoo"
)

var version = "0.4.0"

func main() {
	log.Println("🏛️  Argus — Den med 100 ögon")
	log.Printf("Version %s\n", version)

	ctx := context.Background()

	dbPath := "./data/argus.db"

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatal("❌ Databasfel: ", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		log.Fatal("❌ Schema-fel: ", err)
	}

	// Endast verifierade källor
	sources_ := []sourceInfo{
		{src: sec.New(), flag: "🇺🇸", name: "SEC (EDGAR)", status: "🟢 Verifierad"},
		{src: brreg.New(), flag: "🇳🇴", name: "Brreg (Norge)", status: "🟢 Verifierad"},
		{src: edinet.New(), flag: "🇯🇵", name: "EDINET (Japan)", status: "🟡 API-söks"},
		{src: bolagsverket.New(), flag: "🇸🇪", name: "Bolagsverket", status: "🟡 Nyckel behövs"},
	}

	// Yahoo för kurser
	_ = yahoo.New()

	fmt.Printf("\n✅ Argus är redo!\n\n")
	fmt.Printf("Databas: %s\n", dbPath)
	fmt.Printf("Verifierade register: %d\n\n", len(sources_))

	for _, s := range sources_ {
		fmt.Printf("  %s %s %s — %s\n", s.flag, s.src.Name(), s.name, s.status)
	}
	fmt.Println()
	fmt.Printf("📈 Yahoo Finance — dagliga kurser\n\n")
	fmt.Println("Nästa steg: Börja med SEC EDGAR fetcher.")
}

type sourceInfo struct {
	src    sources.Source
	flag   string
	name   string
	status string
}

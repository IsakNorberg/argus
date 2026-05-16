package main

import (
	"context"
	"fmt"
	"log"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources"
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
	"github.com/IsakNorberg/argus/internal/sources/sec"
	"github.com/IsakNorberg/argus/internal/sources/sedar"
	"github.com/IsakNorberg/argus/internal/sources/yahoo"
	"github.com/IsakNorberg/argus/internal/sources/zefix"
)

var version = "0.3.0"

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

	// Registrera alla källor
	sources_ := []sourceInfo{
		{src: sec.New(), flag: "🇺🇸", name: "SEC (EDGAR)", price: "🟢"},
		{src: bolagsverket.New(), flag: "🇸🇪", name: "Bolagsverket", price: "🟡"},
		{src: brreg.New(), flag: "🇳🇴", name: "Brreg", price: "🟡"},
		{src: cvr.New(), flag: "🇩🇰", name: "CVR", price: "🟢"},
		{src: prh.New(), flag: "🇫🇮", name: "PRH", price: "🟢"},
		{src: companieshouse.New(), flag: "🇬🇧", name: "Companies House", price: "🟢"},
		{src: inpi.New(), flag: "🇫🇷", name: "INPI", price: "🟡"},
		{src: kvk.New(), flag: "🇳🇱", name: "KVK", price: "🟡"},
		{src: edinet.New(), flag: "🇯🇵", name: "EDINET (Japan)", price: "🟢"},
		{src: zefix.New(), flag: "🇨🇭", name: "Zefix (Schweiz)", price: "🟢"},
		{src: cro.New(), flag: "🇮🇪", name: "CRO (Irland)", price: "🟢"},
		{src: sedar.New(), flag: "🇨🇦", name: "SEDAR+ (Kanada)", price: "🟢"},
		{src: ares.New(), flag: "🇨🇿", name: "ARES (Tjeckien)", price: "🟢"},
		{src: krs.New(), flag: "🇵🇱", name: "KRS (Polen)", price: "🟢"},
	}

	// Yahoo för dagens kurs
	_ = yahoo.New()

	fmt.Printf("\n✅ Argus är redo!\n\n")
	fmt.Printf("Databas: %s\n", dbPath)
	fmt.Printf("Register: %d källor\n\n", len(sources_))

	freeCount := 0
	for _, s := range sources_ {
		if s.price == "🟢" {
			freeCount++
			fmt.Printf("  %s %s %s — GRATIS\n", s.flag, s.src.Name(), s.name)
		}
	}
	fmt.Println()
	for _, s := range sources_ {
		if s.price == "🟡" {
			fmt.Printf("  %s %s %s — nyckel/auth\n", s.flag, s.src.Name(), s.name)
		}
	}
	fmt.Println()
	fmt.Printf("📈 Yahoo Finance — dagens kurs\n\n")
	fmt.Printf("Sammanfattning: %d GRATIS av %d register + Yahoo\n", freeCount, len(sources_))
	fmt.Println("\nNästa steg: Implementera fetchers per register.")
}

type sourceInfo struct {
	src  sources.Source
	flag string
	name string
	price string
}

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

var version = "0.4.1"

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

	// Alla register
	sources_ := []sourceInfo{
		{src: sec.New(), flag: "🇺🇸", name: "SEC EDGAR", status: "🟢 Verifierad"},
		{src: brreg.New(), flag: "🇳🇴", name: "Brreg", status: "🟢 Verifierad"},
		{src: bolagsverket.New(), flag: "🇸🇪", name: "Bolagsverket", status: "🟡 Nyckel"},
		{src: edinet.New(), flag: "🇯🇵", name: "EDINET", status: "🟡 API-söks"},
		{src: cvr.New(), flag: "🇩🇰", name: "CVR", status: "❌ 403"},
		{src: prh.New(), flag: "🇫🇮", name: "PRH", status: "❌ 404"},
		{src: companieshouse.New(), flag: "🇬🇧", name: "Companies House", status: "❌ 403"},
		{src: inpi.New(), flag: "🇫🇷", name: "INPI", status: "❌ 403"},
		{src: kvk.New(), flag: "🇳🇱", name: "KVK", status: "❌ Testas"},
		{src: zefix.New(), flag: "🇨🇭", name: "Zefix", status: "❌ 000"},
		{src: cro.New(), flag: "🇮🇪", name: "CRO", status: "❌ 403"},
		{src: sedar.New(), flag: "🇨🇦", name: "SEDAR+", status: "🟡 Sida"},
		{src: ares.New(), flag: "🇨🇿", name: "ARES", status: "❌ HTML SPA"},
		{src: krs.New(), flag: "🇵🇱", name: "KRS", status: "❌ 302"},
	}

	// Yahoo för kurser
	_ = yahoo.New()

	fmt.Printf("\n✅ Argus är redo!\n\n")
	fmt.Printf("Databas: %s\n", dbPath)
	fmt.Printf("Register: %d källor\n\n", len(sources_))

	for _, s := range sources_ {
		fmt.Printf("  %s %-20s — %s\n", s.flag, s.src.Name(), s.status)
	}
	fmt.Println()
	fmt.Printf("📈 Yahoo — dagliga kurser\n\n")
	fmt.Println("Nästa steg: Börja med SEC EDGAR fetcher.")
}

type sourceInfo struct {
	src    sources.Source
	flag   string
	name   string
	status string
}

# Argus Arkitektur

argus/
├── cmd/
│   └── argus/
│       └── main.go          ← Entry point
├── internal/
│   ├── sources/
│   │   ├── source.go        ← Interface för alla källor
│   │   ├── sec/
│   │   │   └── sec.go       ← SEC EDGAR parser
│   │   └── bolagsverket/
│   │       └── bolagsverket.go  ← Bolagsverket parser
│   ├── standardizer/
│   │   └── standardizer.go  ← Normalisering till gemensamt schema
│   └── database/
│       └── database.go      ← SQLite/PostgreSQL
├── pkg/
│   └── models/
│       └── models.go        ← Gemensamma datamodeller
├── docs/
│   └── sources/             ← MD-filer per källa
├── data/                    ← Databas (gitignored)
└── go.mod


## Flöde
1. **Source** hämtar data (SEC, Bolagsverket, etc)
2. **Standardizer** normaliserar till gemensamt format
3. **Database** lagrar

## Data modeller
Alla källor mappar till gemensamma fält i `pkg/models/models.go`.

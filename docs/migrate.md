# Migrering: SQLite → PostgreSQL

## När migrera?
- När flera källor fetchar samtidigt (concurrent writes)
- När du behöver backup/restore
- När du vill köra Argus som en tjänst
- När databasen > 100 000 poster

## Steg-för-steg

### 1. Installera PostgreSQL
```bash
# WSL2 (Ubuntu/Debian)
sudo apt install postgresql postgresql-contrib
sudo service postgresql start

# Eller Docker
docker run -d --name argus-db \
  -e POSTGRES_PASSWORD=argus \
  -e POSTGRES_DB=argus \
  -p 5432:5432 \
  postgres:16
```

### 2. Skapa databas
```bash
sudo -u postgres psql
CREATE DATABASE argus;
CREATE USER argus WITH PASSWORD 'argus';
GRANT ALL PRIVILEGES ON DATABASE argus TO argus;
```

### 3. Byt Go driver
```go
// OLD (sqlite)
import _ "modernc.org/sqlite"
db, _ := sql.Open("sqlite", "data/argus.db")

// NEW (postgres)
import _ "github.com/lib/pq"
db, _ := sql.Open("postgres", "postgres://argus:argus@localhost:5432/argus?sslmode=disable")
```

### 4. Uppdatera connection pool
```go
// PostgreSQL kan hantera många connections
db.SetMaxOpenConns(10)  // SQLite: 1
db.SetMaxIdleConns(5)   // SQLite: 2
```

### 5. SQL-differenser (minimala)

| SQLite | PostgreSQL |
|---|---|
| `AUTOINCREMENT` | `SERIAL` |
| `DATETIME DEFAULT CURRENT_TIMESTAMP` | `TIMESTAMP DEFAULT NOW()` |
| `ON CONFLICT(x) DO UPDATE` | `ON CONFLICT (x) DO UPDATE SET` |
| `INSERT OR REPLACE` | `INSERT ... ON CONFLICT` (samma syntax) |

### 6. Migrera data
```bash
# Export SQLite
sqlite3 data/argus.db ".dump" > export.sql

# Manuell konvertering eller använd verktyg:
# https://github.com/philips-labs/pgloader
pgloader sqlite://data/argus.db postgresql://argus:argus@localhost/argus
```

### 7. Testa
```bash
cd cmd/argus
export ARGUS_DB_DRIVER=postgres
export ARGUS_DB_URL="postgres://argus:argus@localhost:5432/argus?sslmode=disable"
go run .
```

### 8. Backa ut (vid behov)
- Behåll SQLite-koden som fallback
- Använd env-var för att switcha:
```go
driver := os.Getenv("ARGUS_DB_DRIVER")
if driver == "" { driver = "sqlite" }
```

## Tidsuppskattning
- Driver + config: **30 min**
- SQL-anpassningar: **1-2 timmar**
- Testning: **1-2 timmar**
- **Totalt: 4-5 timmar**

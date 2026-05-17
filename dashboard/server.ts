import { serve } from 'bun'
import { Hono } from 'hono'
import { html } from 'hono/html'
import { Database } from 'bun:sqlite'
import { cors } from 'hono/cors'

const app = new Hono()
app.use('/api/*', cors())

// --- SQLite helpers ---
function getDB(): Database | null {
  // Look for DB in parent dir (argus/data/) or local dir
  const paths = [
    Bun.env.ARGUS_DB_PATH,
    '../data/argus.db',
    './data/argus.db',
  ].filter(Boolean) as string[]
  
  for (const path of paths) {
    try {
      return new Database(path, { create: false, read: true, readonly: true })
    } catch { /* try next */ }
  }
  return null
}

function safeQuery<T>(db: Database | null, sql: string): T[] {
  if (!db) return []
  try {
    return db.prepare(sql).all() as T[]
  } catch {
    return []
  }
}

function safeQueryOne<T>(db: Database | null, sql: string): T | null {
  if (!db) return null
  try {
    return db.prepare(sql).get() as T | null
  } catch {
    return null
  }
}

// --- Dashboard overview ---
app.get('/', async (c) => {
  const db = getDB()
  
  const stats = {
    companies: safeQueryOne(db, "SELECT COUNT(*) as n FROM companies") as {n: number} | null,
    financials: safeQueryOne(db, "SELECT COUNT(*) as n FROM financials") as {n: number} | null,
    sources: safeQuery(db, "SELECT source, COUNT(DISTINCT company_id) as companies, MAX(filing_date) as latest FROM financials GROUP BY source ORDER BY source"),
    latestFetches: safeQuery(db, "SELECT source, MAX(filing_date) as last_update FROM financials GROUP BY source ORDER BY last_update DESC LIMIT 10"),
    recentFinancials: safeQuery(db, `
      SELECT f.period, f.source, c.name, c.exchange, f.revenue, f.net_income, f.currency
      FROM financials f JOIN companies c ON f.company_id = c.id
      ORDER BY f.updated_at DESC LIMIT 20
    `),
    exchanges: safeQuery(db, `
      SELECT exchange, COUNT(*) as companies 
      FROM companies 
      WHERE exchange != '' 
      GROUP BY exchange 
      ORDER BY companies DESC
    `),
    topByFinancials: safeQuery(db, `
      SELECT c.name, c.exchange, c.country, COUNT(f.id) as periods 
      FROM companies c JOIN financials f ON c.id = f.company_id 
      GROUP BY c.id 
      ORDER BY periods DESC 
      LIMIT 20
    `) as any[],
  }

  const totalCompan = stats.companies?.n ?? 0
  const totalFin = stats.financials?.n ?? 0

  return c.html(html`
    <!DOCTYPE html>
    <html lang="sv">
    <head>
      <meta charset="utf-8">
      <meta name="viewport" content="width=device-width, initial-scale=1">
      <title>Argus Dashboard</title>
      <script src="https://unpkg.com/htmx.org@1.9.12"></script>
      <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1"></script>
      <style>
        :root {
          --bg: #0d1117;
          --bg2: #161b22;
          --bg3: #21262d;
          --border: #30363d;
          --text: #c9d1d9;
          --text2: #8b949e;
          --accent: #58a6ff;
          --green: #3fb950;
          --yellow: #d29922;
          --red: #f85149;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
          background: var(--bg);
          color: var(--text);
          min-height: 100vh;
        }
        nav {
          background: var(--bg2);
          border-bottom: 1px solid var(--border);
          padding: 1rem 2rem;
          display: flex;
          align-items: center;
          gap: 1rem;
        }
        nav h1 {
          font-size: 1.25rem;
          color: var(--accent);
        }
        nav .tabs {
          display: flex;
          gap: 0.5rem;
          margin-left: auto;
        }
        nav .tab {
          padding: 0.5rem 1rem;
          border-radius: 6px;
          cursor: pointer;
          color: var(--text2);
          border: none;
          background: transparent;
          font-size: 0.9rem;
          transition: all 0.2s;
        }
        nav .tab.active {
          background: var(--bg3);
          color: var(--text);
        }
        nav .tab:hover {
          background: var(--bg3);
        }
        main {
          max-width: 1400px;
          margin: 0 auto;
          padding: 1.5rem;
        }
        .stat-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 1rem;
          margin-bottom: 1.5rem;
        }
        .stat-card {
          background: var(--bg2);
          border: 1px solid var(--border);
          border-radius: 8px;
          padding: 1.25rem;
        }
        .stat-card .value {
          font-size: 2rem;
          font-weight: 700;
          color: var(--accent);
        }
        .stat-card .label {
          color: var(--text2);
          font-size: 0.85rem;
          margin-top: 0.25rem;
        }
        .card {
          background: var(--bg2);
          border: 1px solid var(--border);
          border-radius: 8px;
          margin-bottom: 1rem;
          overflow: hidden;
        }
        .card-header {
          padding: 0.75rem 1rem;
          border-bottom: 1px solid var(--border);
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }
        .card-header h3 { font-size: 0.95rem; }
        table {
          width: 100%;
          border-collapse: collapse;
        }
        th {
          text-align: left;
          padding: 0.5rem 1rem;
          font-size: 0.75rem;
          text-transform: uppercase;
          color: var(--text2);
          border-bottom: 1px solid var(--border);
          background: var(--bg3);
        }
        td {
          padding: 0.5rem 1rem;
          font-size: 0.85rem;
          border-bottom: 1px solid var(--border);
        }
        tr:last-child td { border-bottom: none; }
        tr:hover td { background: var(--bg3); }
        .badge {
          display: inline-block;
          padding: 2px 8px;
          border-radius: 12px;
          font-size: 0.75rem;
          font-weight: 600;
        }
        .badge-source {
          background: var(--bg3);
          color: var(--accent);
          border: 1px solid var(--border);
        }
        .badge-green { background: rgba(63,185,80,0.15); color: var(--green); }
        .badge-yellow { background: rgba(210,153,34,0.15); color: var(--yellow); }
        .exchange-tag {
          color: var(--text2);
          font-size: 0.8rem;
        }
        .two-col {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 1rem;
        }
        @media (max-width: 768px) {
          .two-col { grid-template-columns: 1fr; }
          .stat-grid { grid-template-columns: repeat(2, 1fr); }
          nav { padding: 0.75rem 1rem; }
          main { padding: 1rem; }
        }
      </style>
    </head>
    <body>
      <nav>
        <h1>⚡ Argus</h1>
        <span style="color: var(--text2); font-size: 0.85rem;">v0.4.3-dashboard</span>
        <div class="tabs">
          <button class="tab active" hx-swap="none" onclick="showTab('overview')">Översikt</button>
          <button class="tab" onclick="showTab('sources')">Källor</button>
          <button class="tab" onclick="showTab('data')">Data</button>
        </div>
      </nav>
      <main>
        <!-- Overview stat cards -->
        <div class="stat-grid">
          <div class="stat-card">
            <div class="value">${totalCompan.toLocaleString()}</div>
            <div class="label">Bolag i databasen</div>
          </div>
          <div class="stat-card">
            <div class="value">${totalFin.toLocaleString()}</div>
            <div class="label">Finansiella perioder</div>
          </div>
          <div class="stat-card">
            <div class="value">${stats.exchanges.length}</div>
            <div class="label">Aktiva börser</div>
          </div>
          <div class="stat-card">
            <div class="value">${stats.sources.length}</div>
            <div class="label">Datakällor</div>
          </div>
        </div>

        <!-- Exchanges -->
        <div class="card">
          <div class="card-header">
            <h3>📊 Börser</h3>
          </div>
          <table>
            <thead>
              <tr>
                <th>Börs</th>
                <th>Bolag</th>
              </tr>
            </thead>
            <tbody>
              ${stats.exchanges.map((e: any) => html`
                <tr>
                  <td><strong>${e.exchange}</strong></td>
                  <td>${e.companies}</td>
                </tr>
              `)}
              ${stats.exchanges.length === 0 ? html`<tr><td colspan="2" style="color: var(--text2); text-align: center; padding: 1rem;">Inga börsregistreringar ännu. Kör <code>argus fetch --all</code></td></tr>` : ''}
            </tbody>
          </table>
        </div>

        <div class="two-col">
          <!-- Källstatus -->
          <div class="card">
            <div class="card-header">
              <h3>🏛️ Datakällor</h3>
            </div>
            <table>
              <thead>
                <tr>
                  <th>Källa</th>
                  <th>Bolag</th>
                  <th>Senaste</th>
                </tr>
              </thead>
              <tbody>
                ${stats.sources.map((s: any) => html`
                  <tr>
                    <td><span class="badge badge-source">${s.source}</span></td>
                    <td>${s.companies}</td>
                    <td>${s.latest ? new Date(s.latest).toLocaleDateString('sv-SE') : '—'}</td>
                  </tr>
                `)}
                ${stats.sources.length === 0 ? html`<tr><td colspan="3" style="color: var(--text2); text-align: center; padding: 1rem;">Ingen data ännu</td></tr>` : ''}
              </tbody>
            </table>
          </div>

          <!-- Senaste finansiella -->
          <div class="card">
            <div class="card-header">
              <h3>📈 Senaste finansiella data</h3>
            </div>
            <table>
              <thead>
                <tr>
                  <th>Bolag</th>
                  <th>Period</th>
                  <th>Intäkt</th>
                  <th>Vinst</th>
                </tr>
              </thead>
              <tbody>
                ${stats.recentFinancials.map((f: any) => html`
                  <tr>
                    <td>${f.name} <span class="exchange-tag">(${f.exchange})</span></td>
                    <td>${f.period}</td>
                    <td>${f.revenue ? (f.revenue / 1e6).toFixed(1) + 'M ' + f.currency : '—'}</td>
                    <td style="color: ${parseFloat(f.net_income) >= 0 ? 'var(--green)' : 'var(--red)'}">
                      ${f.net_income ? (f.net_income / 1e6).toFixed(1) + 'M ' + f.currency : '—'}
                    </td>
                  </tr>
                `)}
                ${stats.recentFinancials.length === 0 ? html`<tr><td colspan="4" style="color: var(--text2); text-align: center; padding: 1rem;">Ingen finansiell data</td></tr>` : ''}
              </tbody>
            </table>
          </div>
        </div>

        <!-- Top bolag -->
        <div class="card">
          <div class="card-header">
            <h3>🏆 Bolag med mest data</h3>
          </div>
          <table>
            <thead>
              <tr>
                <th>Bolag</th>
                <th>Börs</th>
                <th>Land</th>
                <th>Perioder</th>
              </tr>
            </thead>
            <tbody>
              ${stats.topByFinancials.map((c: any) => html`
                <tr>
                  <td><strong>${c.name}</strong></td>
                  <td>${c.exchange}</td>
                  <td>${c.country}</td>
                  <td><span class="badge badge-green">${c.periods}</span></td>
                </tr>
              `)}
              ${stats.topByFinancials.length === 0 ? html`<tr><td colspan="4" style="color: var(--text2); text-align: center; padding: 1rem;">Inga resultat</td></tr>` : ''}
            </tbody>
          </table>
        </div>
      </main>

      <script>
        function showTab(name) {
          document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
          event.target.classList.add('active');
        }
      </script>
    </body>
    </html>
  `)
})

// --- API endpoints ---
app.get('/api/status', (c) => {
  const db = getDB()
  return c.json({
    companies: (safeQueryOne(db, "SELECT COUNT(*) as n FROM companies") as {n: number} | null)?.n ?? 0,
    financials: (safeQueryOne(db, "SELECT COUNT(*) as n FROM financials") as {n: number} | null)?.n ?? 0,
    sources: safeQuery(db, "SELECT source, COUNT(DISTINCT company_id) as count FROM financials GROUP BY source"),
    exchanges: safeQuery(db, "SELECT exchange, COUNT(*) as count FROM companies WHERE exchange != '' GROUP BY exchange ORDER BY count DESC"),
  })
})

app.get('/api/companies', (c) => {
  const db = getDB()
  const limit = c.req.query('limit') || '50'
  return c.json(safeQuery(db, `
    SELECT name, ticker, exchange, country, industry, created_at
    FROM companies
    ORDER BY created_at DESC
    LIMIT ${limit}
  `))
})

// --- Server ---
const port = parseInt(Bun.env.ARGUS_DASHBOARD_PORT || '8001')
console.log(`⚡ Argus Dashboard på http://localhost:${port}`)
serve({ port, fetch: app.fetch })

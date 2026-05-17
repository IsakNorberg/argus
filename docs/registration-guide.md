# Registreringsguide — API Key & Access

> Hur du skaffar access till de register som kraver registrering.
> Uppdaterad: 2026-05-16

---

## 1. Companies House (UK) 🇬🇧 — GRATIS

**Tid:** ~5 minuter
**Kostnad:** GRATIS
**Rate limit:** 600 requests / 5 min

### Steg-för-steg

1. Gå till: https://developer.company-information.service.gov.uk/
2. Klicka på **"Sign in"** → **"Register"**
3. Skapa konto:
   - **Email** — din vanliga email
   - **Lösenord** — valfritt
   - **Namn** — fullständigt namn
4. Verifiera email (kolla inbox/spam)
5. Logga in och gå till **"My applications"**
6. Klicka **"Create new application"**
   - **Name:** `Argus` eller valfritt
   - **Description:** `Data aggregation for financial research`
7. Du får en **API key** (lång sträng)
8. Spara nyckeln!

### Testa direkt

```bash
# Ersätt YOUR_API_KEY med din nyckel
curl -u "YOUR_API_KEY:" \
  "https://api.company-information.service.gov.uk/search/companies?q=test"
```

### Använda i Argus

```bash
# Spara i .env
export COMPANIES_HOUSE_API_KEY="din_nyckel_har"
```

---

## 2. EDINET (Japan) 🇯🇵 — GRATIS efter registrering

**Tid:** ~15-30 minuter
**Kostnad:** GRATIS
**Obs:** Kräver Microsoft-konto (eller arbets/skolkonto)

### Steg-för-steg

1. Gå till: https://disclosure2.edinet-fsa.go.jp/
2. Klicka på **"EDINET API"** eller **"EDINET APIの利用登録はこちらから"**
3. Klicka **"ログイン"** (Login)
4. Du omdirigeras till **Microsoft inloggning**:
   - Använd ett **Microsoft-konto** (Outlook, Hotmail, live.com)
   - Eller **arbets-/skolkonto**
   - Skapa nytt Microsoft-konto om du saknar: https://account.microsoft.com/account
5. Godkänn behörigheter
6. Efter inloggning:
   - Du ska få tillgång till API-dokumentation
   - Eventuellt en API-nyckel eller session-token
7. Hitta API-dokumentation (sök efter "API 仕様" eller "仕様書")

### Om du inte hittar API-key direkt

EDINET använder **session cookies** istället för API-nycklar:
- Logga in via webbläsaren
- Extrahera cookie-värden: `EDINET_SessionId`, `GX_SESSION_ID`
- Använd dessa cookies i dina requests

### Alternativt: Använd XBRL-filer direkt

EDINET erbjuder XBRL-filer för nedladdning utan API:
- Gå till: https://disclosure2.edinet-fsa.go.jp/WEEK0010.aspx
- Ladda ner taxonomy-filer
- Använd dessa för att mappa fält

---

## 3. Bolagsverket (Sverige) 🇸🇪 — API-nyckel krävs

**Tid:** 1-14 dagar (ansökan)
**Kostnad:** Varierar
**Obs:** Officiell process kan ta tid

### Steg-för-steg

1. Gå till: https://bolagsverket.se
2. Sök efter **"API"** eller **"Öppna data"**
3. Alternativt: https://bolagsverket.se/tjanster/oppnadata
4. Fyll i ansökan:
   - **Organisation** (företag/privatperson)
   - **Syfte** — "Finansiell data-aggregering för research"
   - **Användningsområde** — API-access
5. Vänta på godkännande
6. Du får en API-nyckel via email

### Alternativ: Third-party

Om Bolagsverket är för långsamt, kolla:
- https://allabolag.se — har API via Bolagsverket
- https://merinfo.se — alternativ datakälla

---

## 4. CVR (Danmark) 🇩🇰 — Via Datafordeler

**Tid:** ~30 minuter
**Kostnad:** GRATIS (CVR-data)
**Obs:** Kräver NemID/MitID

### Steg-för-steg

1. Gå till: https://datafordeler.dk
2. Klicka **"Log ind"**
3. **Kräver NemID/MitID** (dansk digital ID)
4. Efter inloggning:
   - Sök efter **"CVR"** eller **"Virksomheder"**
   - Välj datakälla
5. Du får:
   - GraphQL-access
   - File download
   - OAuth credentials
6. För API-access:
   - Skapa en "Data abonnement"
   - Välj CVR-datasættet
   - Få API-nyckel

### Utan NemID

Om du saknar NemID, använd **third-party**:
- https://cvrapi.dk — Enkel API, gratis provperiod
- Registrera:https://cvrapi.dk/signup
- Du får en API key direkt

---

## 5. INPI (Frankrike) 🇫🇷 — Cloudflare block

**Tid:** Okänd
**Kostnad:** Varierar
**Obs:** Cloudflare blockerar server-requests

### Alternativ 1: Sirene API (via api.gouv.fr)

1. Gå till: https://api.entreprise.api.gouv.fr/
2. Kräver troligen **fransk organisation** eller EU-server
3. Om DNS fail kvarstår — testa från en fransk VPS

### Alternativ 2: data.inpi.fr

1. Gå till: https://data.inpi.fr/
2. Klicka **"API et données"**
3. Registrera konto
4. Du kan få:
   - Bulk data-dumpar
   - API-access efter godkännande

---

## 6. KVK (Nederländerna) 🇳🇱 — GRATIS test-API redan!

**Ingen registrering krävs för test-API!**

Redan fungerande key:
```
l7xx1f2691f2520d487b902f4e0b57a0b197
```

### Testa direkt

```bash
curl "https://api.kvk.nl/test/api/v2/zoeken?naam=test" \
  -H "apikey: l7xx1f2691f2520d487b902f4e0b57a0b197" \
  -H "accept: application/json"
```

### Production API (valfritt)

1. Gå till: https://developers.kvk.nl/nl/
2. Klicka **"Abonnement aanvragen"**
3. Välj paket (månatlig kostnad)
4. Du får production API-nyckel

---

## Sammanfattning — Prioriterad lista

| # | Register | Tid | Kostnad | Action |
|---|---|---|---|---|
| 1 | **Companies House UK** | 5 min | GRATIS | **Gör detta först** — enkelt |
| 2 | **KVK NL** | 0 min | GRATIS | ⚠️ Test-API redan fungerar! |
| 3 | **EDINET Japan** | 30 min | GRATIS | Efter UK—kräver Microsoft-konto |
| 4 | **CVR DK** | 30 min | GRATIS | Kräver NemID — svårt utomlands |
| 5 | **INPI FR** | Okänd | Varierar | Avvakta—svårt |
| 6 | **Bolagsverket SE** | 1-14 dagar | Varierar | Ansök—lång väntetid |

---

## Rekommendation

1. **Gör Companies House NU** — 5 minuter, gratis, fungerar direkt
2. **KVK** — redan fungerar, ingen action needed
3. **EDINET** — om du har Microsoft-konto, gör detta
4. **Övriga** — vänta, fokusera på att bygga SEC + PRH först

# Datakällor — Pris och Tillgänglighet

## ✅ Helt Gratis (publikt API, ingen nyckel)

| Register | Land | API | Format | Begränsning |
|---|---|---|---|---|
| SEC EDGAR | 🇺🇸 USA | REST | XBRL, JSON | User-Agent med email |
| CVR / Erhvervsstyrelsen | 🇩🇰 Danmark | REST | XBRL, XML | 250 anrop/min |
| Companies House | 🇬🇧 UK | REST | JSON, PDF, iXBRL | 600 anrop/5min |
| PRH (YTJ) | 🇫🇮 Finland | REST | JSON | Publikt |

## 🔑 Gratis men kräver registrering

| Register | Land | API | Format | Begränsning |
|---|---|---|---|---|
| Brønnøysundregistrene | 🇳🇴 Norge | REST | JSON, XML | Grundläggande auth |
| INPI / data.inpi.fr | 🇫🇷 Frankrike | REST | JSON | Publikt, User-Agent |
| Bolagsverket | 🇸🇪 Sverige | REST | JSON | API-nyckel behövs |
| KVK | 🇳🇱 Nederländerna | REST | JSON | API-nyckel behövs |

## ❌ Begränsad/Betald

| Register | Land | Status |
|---|---|---|
| Bundesanzeiger | 🇩🇪 Tyskland | PDF, ingen öppen API. Många endpoints betalda |
| ESEF | 🇪🇺 EU | XBRL-filer finns, men ingen "API" i vanlig mening |
| Registro Mercantil | 🇪🇸 Spanien | PDF-arkiv. Mycket begränsad öppen data |
| Yahoo Finance | 🌍 Global | Oofficiell API — kan sluta fungera. Endast kurser |

## 📋 Prioriterad ordning (börja med gratis)

1. **SEC EDGAR** 🇺🇸 — Bäst API, XBRL, helt gratis, ingen nyckel
2. **CVR (Danmark)** 🇩🇰 — REST API, helt gratis, XBRL
3. **Companies House (UK)** 🇬🇧 — REST API, helt gratis, iXBRL
4. **PRH (Finland)** 🇫🇮 — REST API, gratis
5. **Brreg (Norge)** 🇳🇴 — REST API, enkel auth
6. **INPI (Frankrike)** 🇫🇷 — REST API, gratis

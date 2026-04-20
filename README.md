# SecScan — Web Security Scanner

> **Demo URL:** _coming after F10 deploy_  
> **Stack:** Go 1.22 + Gin · Next.js 14 + TypeScript + TailwindCSS · Docker Compose

SecScan takes a URL and runs **7 security modules in parallel**, returning a comprehensive A/F-grade report.

---

## Quick Start

```bash
# Clone & run everything
git clone https://github.com/<username>/secscan
cd secscan
docker compose up --build

# Frontend → http://localhost:3000
# Backend  → http://localhost:8080/health
```

## Scanner Modules

| Module | Description | Status |
|--------|-------------|--------|
| `ports` | TCP port scan + service detection | 🔄 F03 |
| `headers` | Security header audit + grade | 🔄 F04 |
| `tls` | TLS version + cipher + cert check | 🔄 F05 |
| `fuzz` | Directory/path fuzzing | 🔄 F06 |
| `xss` | XSS payload reflection check | 🔄 F06 |
| `sqli` | Error/boolean/time-based detection | 🔄 F06 |
| `cve` | Tech detection + NVD/OSV lookup | 🔄 F07 |

## API

```
POST /api/scan              → { scan_id, status }
GET  /api/scan/:id          → Full JSON report
GET  /api/scan/:id/stream   → SSE live progress
GET  /api/scan/:id/report.pdf → PDF download
GET  /health                → { status: "ok" }
```

## Security Design

- **SSRF Protection**: All URLs DNS-resolved and checked against private CIDR ranges before scanning
- **Semgrep SAST**: Runs on every PR (OWASP Top 10 + Go rules)
- **Trivy**: Weekly CVE scan on filesystem + images

## AI Kullanım Beyanı

Bu projede yapay zeka asistan kullanıldı. Tüm kod mantığı anlaşılmış ve açıklanabilir düzeyde geliştirilmiştir.

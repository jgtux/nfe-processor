# NF-e Processor

Full Stack application for receiving, processing and classifying NF-e (Nota Fiscal Eletrônica) XML files.

---

## Stack

| Layer     | Technology                               |
|-----------|------------------------------------------|
| Backend   | Go 1.22 + Gin                            |
| Queue     | RabbitMQ 3.13 (persistent messages)      |
| Database  | PostgreSQL 16                            |
| Frontend  | Vue 3 + Vite + TypeScript + Tailwind CSS |
| Container | Docker + Docker Compose                  |
| API Docs  | Swagger / OpenAPI (swag)                 |

---

## Architecture

```
┌─────────────┐     POST /xml/upload      ┌────────────────┐
│  Vue 3 SPA  │ ─────────────────────────▶│  Gin HTTP API  │
└─────────────┘                           └───────┬────────┘
                                                  │ SaveUpload (PostgreSQL)
                                                  │ Publish (UploadID only)
                                          ┌───────▼────────┐
                                          │   RabbitMQ     │  ← durable queue
                                          └───────┬────────┘
                                                  │ Consume
                                          ┌───────▼────────┐
                                          │    Consumer    │
                                          │  (goroutine)   │
                                          │ fetch → parse  │
                                          │ classify →     │
                                          │ persist        │
                                          └───────┬────────┘
                                                  │
                                          ┌───────▼────────┐
                                          │  PostgreSQL 16 │
                                          │  nfe_uploads   │
                                          │  nfes          │
                                          └────────────────┘
```

### Processing flow

1. **Upload** — Frontend sends one or more XMLs via `POST /api/v1/xml/upload`
2. **Persist** — Raw XML is saved to `nfe_uploads` table in PostgreSQL
3. **Enqueue** — Only the `upload_id` (UUID) is published to RabbitMQ — no XML in the queue
4. **Consume** — A goroutine consumes the queue:
   - Fetches the raw XML from PostgreSQL by `upload_id`
   - Validates against the official `procNFe_v4.00.xsd` schema
   - Parses the XML (access key, CNPJ, issuer, recipient, amount, date)
   - Validates access key mod 11 check digit and issuer CNPJ mod 11
   - Queries the internal clients mock API
   - Classifies the operation: **Purchase** / **Sale** / **Unidentified**
   - Persists the result to `nfes` table
   - Deletes raw XML from `nfe_uploads` (data is now in `nfes`)
5. **Query** — Frontend fetches processed data via REST API

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with **buildx plugin**
- [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.22+](https://golang.org/dl/) — required to generate Swagger docs locally
- [swag CLI](https://github.com/swaggo/swag) — required to generate Swagger docs

### Installing prerequisites

**Arch / Artix**
```bash
sudo pacman -S go nodejs npm docker docker-compose docker-buildx
```

**macOS**
```bash
brew install go node docker
```

**Ubuntu / Debian**
```bash
sudo apt install golang nodejs npm docker.io docker-compose-plugin
```

### Enabling Docker

**systemd** (Ubuntu, Debian, Fedora, Arch, etc.)
```bash
sudo systemctl enable --now docker
```

**OpenRC** (Artix, Gentoo)
```bash
sudo rc-update add docker default
sudo rc-service docker start
```

**runit** (Artix, Void)
```bash
sudo ln -s /etc/runit/sv/docker /run/runit/service/
```

**dinit** (Artix)
```bash
sudo dinitctl enable docker
sudo dinitctl start docker
```

Add your user to the docker group (all init systems):
```bash
sudo usermod -aG docker $USER
newgrp docker
```

Enable buildx:
```bash
docker buildx install
```

---

## First-time setup

**1. Install swag and generate Swagger docs**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
cd backend
swag init -g cmd/server/main.go -o docs
cd ..
```

**2. Copy environment file**
```bash
cp .env.example .env
```

> `go.sum` and `package-lock.json` are already committed to the repository.
> `go mod tidy` and `npm install` are not required before the first build.

---

## Running

### All services via Docker Compose
```bash
make up
```

| Service          | URL                                      |
|------------------|------------------------------------------|
| Frontend         | http://localhost:3000                    |
| API              | http://localhost:8080/api/v1             |
| Swagger UI       | http://localhost:8080/swagger/index.html |
| RabbitMQ Manager | http://localhost:15672  (guest / guest)  |

### Stop services
```bash
make down
```

### Rebuild images
```bash
make build
```

### Follow logs
```bash
make logs
make logs-backend
```

---

## Local development (without Docker)

Start infrastructure only:
```bash
docker compose up postgres rabbitmq -d
```

Run backend locally:
```bash
make backend-dev
```

Run frontend dev server:
```bash
make frontend-dev
# http://localhost:5173
```

---

## Makefile reference

| Command                 | Description                           |
|-------------------------|---------------------------------------|
| `make up`               | Start all services                    |
| `make down`             | Stop all services                     |
| `make build`            | Rebuild Docker images                 |
| `make logs`             | Follow all logs                       |
| `make logs-backend`     | Follow backend logs only              |
| `make backend-dev`      | Run backend locally                   |
| `make backend-tidy`     | go mod tidy                           |
| `make backend-swagger`  | Regenerate Swagger docs               |
| `make backend-test`     | Run tests                             |
| `make frontend-install` | npm install                           |
| `make frontend-dev`     | Vite dev server                       |
| `make frontend-build`   | Production build                      |
| `make db-shell`         | psql into PostgreSQL container        |
| `make mq-ui`            | Open RabbitMQ management UI           |
| `make lint`             | go vet + tsc                          |
| `make clean`            | Remove containers, volumes, artifacts |

---

## API Endpoints

| Method | Endpoint                   | Description                                        |
|--------|----------------------------|----------------------------------------------------|
| POST   | `/api/v1/xml/upload`       | Upload one or more XML files (max 10, max 1MB each)|
| GET    | `/api/v1/nfe`              | List all processed NF-es (excludes quarantine)     |
| GET    | `/api/v1/nfe/summary`      | Purchase/sale count per client                     |
| GET    | `/api/v1/nfe/unidentified` | NF-es with no internal client match                |
| GET    | `/api/v1/nfe/quarantine`   | NF-es that failed validation or parsing            |
| GET    | `/api/v1/clients`          | List internal clients (mock)                       |
| GET    | `/api/v1/health`           | Health check                                       |
| GET    | `/swagger/index.html`      | Interactive API documentation                      |

---

## Usage examples

```bash
# Upload XML files
curl -X POST http://localhost:8080/api/v1/xml/upload \
  -F "files=@examples/nfe_compra.xml" \
  -F "files=@examples/nfe_venda.xml" \
  -F "files=@examples/nfe_nao_identificada.xml"

# Wait ~1s then query results
curl http://localhost:8080/api/v1/nfe | jq .

# Client summary
curl http://localhost:8080/api/v1/nfe/summary | jq .

# Unidentified NF-es
curl http://localhost:8080/api/v1/nfe/unidentified | jq .

# Quarantine (failed uploads)
curl http://localhost:8080/api/v1/nfe/quarantine | jq .
```

---

## Smoke tests

```bash
cd tests/smoke
pip install requests
python3 smoke_test.py

# Against a different host
python3 smoke_test.py --base-url http://192.168.1.100:8080
```

Covers: health, clients, upload, async processing, classification, summary,
unidentified, quarantine, file size limit, file count limit, non-XML rejection,
and rate limiting.

---

## Internal clients (mock)

XMLs whose issuer or recipient CNPJ matches one of the following will be identified.
All CNPJs are mathematically valid (mod 11 verified).

| Name                    | CNPJ           |
|-------------------------|----------------|
| Empresa Alpha Ltda      | 10433218000193 |
| Empresa Beta S.A.       | 19600133000127 |
| Comércio Gama Eireli    | 89083863000183 |
| Distribuidora Delta ME  | 79402654000100 |
| Indústria Épsilon Ltda  | 23511615000188 |

---

## Project structure

```
nfe-processor/
├── backend/
│   ├── cmd/server/main.go              # Entrypoint — wires all layers together
│   ├── internal/
│   │   ├── config/                     # Environment variable loading
│   │   ├── domain/                     # Domain types (NFe, OperationType, ...)
│   │   ├── parser/
│   │   │   ├── parser.go               # NF-e XML parser (nfeProc only)
│   │   │   ├── validate.go             # Mod 11 (access key + CNPJ) validation
│   │   │   ├── xsd.go                  # XSD structural validation via libxml2
│   │   │   └── schemas/                # Official SEFAZ XSD schemas (see below)
│   │   ├── queue/                      # RabbitMQ producer and consumer
│   │   ├── mock/                       # Internal clients mock
│   │   ├── repository/                 # PostgreSQL access (sqlx)
│   │   ├── service/                    # Business logic and orchestration
│   │   └── api/
│   │       ├── handler/                # HTTP handlers + generic Response[T]
│   │       └── middleware/             # Request logger + rate limiter
│   └── docs/                           # Generated by swag
├── frontend/
│   └── src/
│       ├── pages/                      # Upload, Dashboard, Unidentified, Quarantine
│       ├── components/                 # StatCard, OperationBadge, StatusBadge
│       ├── services/api.ts             # API communication layer
│       ├── types/index.ts              # TypeScript types
│       └── router/                     # Vue Router
├── examples/                           # Sample XML files for testing
├── tests/
│   └── smoke/smoke_test.py             # End-to-end smoke tests
├── Makefile
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Technical decisions

- **Raw XML stored in PostgreSQL** — `nfe_uploads` table holds the original XML as `BYTEA`. The queue carries only the `upload_id`, keeping RabbitMQ lightweight. After processing, the raw XML is deleted — data lives in `nfes`.
- **Durable RabbitMQ messages** — `durable: true` on the queue and `DeliveryMode: Persistent` on messages ensure no upload is lost on broker restart.
- **Nack with requeue** — Transient errors (e.g. database unavailable) requeue the message. Permanent errors (invalid XML) are acked and persisted with `error` status in quarantine.
- **QoS = 1** — Each worker processes one message at a time for fair dispatch.
- **Upsert by access key** — Reprocessing the same XML is idempotent.
- **No ORM** — `sqlx` was chosen for simplicity and full control over queries.
- **Generic `Response[T]`** — All endpoints share a single typed envelope, making the API contract consistent and easy to extend.
- **Quarantine with TTL** — NF-es that fail XSD validation, mod 11 checks, or parsing are isolated from the main dashboard and automatically deleted after `QUARANTINE_TTL_DAYS` days. Quarantine persist errors trigger nack so no failure goes unrecorded.
- **Rate limiter** — Token bucket algorithm per client IP on the upload endpoint. No external dependencies — implemented with `sync.Map`. Configurable via `RATE_LIMIT_BURST` and `RATE_LIMIT_RPS`.
- **Upload hardening** — Requests are limited to 10 files and 1 MB per file. A content pre-check (`<?xml` / `<nfeProc`) rejects obvious non-XML before enqueuing.
- **Mod 11 validation** — Both the NF-e access key (SEFAZ MOC section 5.4, weights 2–9) and the issuer CNPJ (Receita Federal, weights 5,4,3,2,9,8,7,6,5,4,3,2) are validated before persistence.
- **XSD structural validation only** — Cryptographic signature verification (XMLDSig) requires C14N canonicalization, which has no mature native support in Go. Structural integrity is validated; signature authenticity is a known limitation.

---

## XSD Schema files

The `backend/internal/parser/schemas/` directory contains the official NF-e v4.00 XSD schemas.
Only authorized NF-e documents in `<nfeProc>` format are accepted.

**Required files and their dependency chain:**

| File | Purpose |
|---|---|
| `procNFe_v4.00.xsd` | Entry point — defines the `<nfeProc>` root element |
| `leiauteNFe_v4.00.xsd` | Full NF-e fiscal layout (included by procNFe) |
| `tiposBasico_v4.00.xsd` | Basic fiscal types (included by leiauteNFe) |
| `xmldsig-core-schema_v1.01.xsd` | XML digital signature types, ICP-Brasil profile (imported by leiauteNFe) |

**Source:** All files are published by SEFAZ/ENCAT and mirrored at
`https://github.com/nfephp-org/sped-nfe/tree/master/schemes/PL_009_V4`

The `xmldsig-core-schema_v1.01.xsd` is a SEFAZ-specific profile of the W3C XMLDSig schema
that fixes the signature algorithm to `rsa-sha1` and digest to `sha1`, as required by ICP-Brasil.

---

## Notes

### Environment variables

All variables have fallback values defined in `config.go`, so the application starts without a `.env`
file. However, the fallbacks use weak credentials and point to `localhost`.

**Never run with fallback values in production.** Always provide all variables explicitly:

| Variable                          | Fallback                               | Description                                             |
|-----------------------------------|----------------------------------------|---------------------------------------------------------|
| `DB_HOST`                         | `localhost`                            | PostgreSQL host                                         |
| `DB_PORT`                         | `5432`                                 | PostgreSQL port                                         |
| `DB_USER`                         | `nfe`                                  | PostgreSQL user                                         |
| `DB_PASSWORD`                     | `nfe123` ⚠️                            | PostgreSQL password                                     |
| `DB_NAME`                         | `nfe_db`                               | PostgreSQL database name                                |
| `DB_SSLMODE`                      | `disable`                              | PostgreSQL SSL mode                                     |
| `RABBITMQ_URL`                    | `amqp://guest:guest@localhost:5672/` ⚠️| RabbitMQ connection URL                                 |
| `RABBITMQ_QUEUE`                  | `nfe_queue`                            | Queue name                                              |
| `SERVER_PORT`                     | `8080`                                 | HTTP server port                                        |
| `QUARANTINE_TTL_DAYS`             | `30`                                   | Days before quarantined NF-es are automatically deleted |
| `QUARANTINE_CLEANUP_INTERVAL_HOURS` | `24`                                 | Hours between cleanup runs                              |
| `RATE_LIMIT_BURST`                | `10`                                   | Upload endpoint burst size (token bucket)               |
| `RATE_LIMIT_RPS`                  | `2`                                    | Upload endpoint sustained rate (requests per second)    |

### Frontend environment

The frontend has no `.env` file. The API URL is resolved via proxy in both environments:

- **Development** — Vite proxies `/api` to `http://localhost:8080` (configured in `vite.config.ts`)
- **Production** — Nginx proxies `/api` to `http://backend:8080` (configured in `nginx.conf`)

If the backend host or port changes in production, `nginx.conf` must be updated and the frontend
image rebuilt. A `VITE_API_BASE_URL` variable with `envsubst` in the Nginx entrypoint would make
this configurable without a rebuild, but is out of scope for this project.

### Reproducibility

Docker image tags (`postgres:16-alpine`, `rabbitmq:3.13-management-alpine`, etc.) are pinned to a
major/minor version but not to a specific digest. For a fully hermetic build in production,
replace tags with their SHA256 digest:

```
FROM golang:1.22-alpine@sha256:<hash>
```

Go dependencies are fully locked via `go.sum`. Node dependencies are locked via `package-lock.json`
with exact version pins in `package.json`.

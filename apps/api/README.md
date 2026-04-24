# API (Go)

HTTP API for Finance Insights. Auth is not implemented yet.

## Requirements

- Go 1.25
- PostgreSQL 14+
- `pdftotext` from [poppler-utils](https://poppler.freedesktop.org/) on `PATH` (used by the PDF parser). On macOS: `brew install poppler`. On Debian/Ubuntu: `apt-get install poppler-utils`.

## Configuration

The server reads configuration from environment variables.

| Variable        | Required | Default   | Description                                            |
| --------------- | -------- | --------- | ------------------------------------------------------ |
| `HTTP_ADDR`     | no       | `:8080`   | Address the HTTP server binds to.                      |
| `DATABASE_URL`  | yes      | —         | PostgreSQL connection string (pgx format).             |
| `STORAGE_DIR`   | no       | `storage` | Root directory for uploaded statements and outputs.    |

Example:

```bash
export DATABASE_URL="postgres://finance:finance@localhost:5432/finance_insights?sslmode=disable"
export STORAGE_DIR="../../storage"
```

## Run locally

```bash
cd apps/api
go run ./cmd/server
```

On startup the server:

1. opens a pgx connection pool against `DATABASE_URL`
2. applies any pending SQL migrations embedded under `internal/db/migrations/`
3. serves HTTP traffic on `HTTP_ADDR`

## Migrations

Migrations are plain `.sql` files in `internal/db/migrations/`, applied in lexical order. Each migration is executed in its own transaction, and applied versions are recorded in a `schema_migrations` table. To add a new migration, create a new file following the `NNNN_description.sql` convention.

## Endpoints

### `GET /health`

Liveness probe.

```bash
curl -s http://localhost:8080/health
# {"status":"ok"}
```

### `POST /statements`

Upload a bank statement PDF. The file is saved under `$STORAGE_DIR/statements/<id>.pdf`, parsed, and its transactions persisted atomically.

- Request: `multipart/form-data` with a `file` field containing the PDF.
- Max upload size: 25 MiB.
- Accepts `application/pdf` (or `application/octet-stream`) with a `.pdf` filename.

```bash
curl -s -X POST http://localhost:8080/statements \
  -F "file=@./sample.pdf"
```

Response `201 Created`:

```json
{
  "id": "0190ec3d-...",
  "fileName": "sample.pdf",
  "transactionCount": 42
}
```

### `GET /transactions`

List transactions ordered by `date DESC, id DESC`.

Query parameters:

- `limit` (optional, default `50`, max `200`)
- `offset` (optional, default `0`)

```bash
curl -s "http://localhost:8080/transactions?limit=20&offset=0"
```

Response `200 OK`:

```json
{
  "items": [
    {
      "id": "0190ec3e-...",
      "statementId": "0190ec3d-...",
      "accountId": "5009-112563-658",
      "date": "2024-07-08",
      "merchantIdentifier": "STARBUCKS",
      "description": "CARD PURCHASE",
      "direction": "Debit",
      "amount": "71.49",
      "balanceAfterTransaction": "1987.26",
      "mccCode": "5411",
      "cardMasked": "516732xxxxxx4321",
      "reference": "EL01P 0442174",
      "bankReferenceNumber": "2960",
      "paymentMethod": "GOOGLE-PAY"
    }
  ],
  "total": 142,
  "limit": 20,
  "offset": 0
}
```

Amounts are returned as strings to preserve exact `numeric(14,2)` precision.

## Testing

```bash
go test ./...
```

The current suite covers handlers and services with in-package fakes. Repository and end-to-end ingest tests (which require a live Postgres) are not yet included.

# AGENTS.md

## Scope
This file applies to everything under `apps/api`.

## Tech
- Go 1.25
- PostgreSQL
- Docker Compose

## Backend architecture
Use a layered structure:
- handler/router layer → HTTP transport only
- service layer → business logic
- repository layer → persistence
- parser layer → statement parsing per bank and version

Preferred flow:
`handler -> service -> repository`

## Package layout

```
internal/
├── domain/          ← one file per entity; no imports from our packages
├── parsers/         ← BankParser interface, ParsedTransaction, Registry
│   └── {bank}/
│       └── {version}/   ← e.g. piraeus/v1/
├── repository/      ← SQL queries only; imports domain
├── service/         ← business logic; imports domain and parsers
├── handler/         ← HTTP transport; imports domain and service
└── server/          ← router registration
```

## Domain rules
- All domain entities live in `internal/domain/`, one file per entity.
- Domain types are plain Go structs with no methods or database tags.
- Repositories, services, and handlers all import `domain` for shared types.
- Never define entity structs outside of the `domain` package.

## Repository rules
- SELECT, INSERT, UPDATE, DELETE queries belong in repository files.
- Never put SELECT queries inside migration files — migrations are DDL only.
- Repository methods take and return domain types.
- No business logic in repositories.

## Parser rules
- Parsers are organized as `internal/parsers/{bank-name}/{version}/`.
- Each versioned parser implements the `parsers.BankParser` interface:
  `BankName() string`, `Version() string`, `Parse(pdfPath string) ([]ParsedTransaction, error)`.
- Register parsers in `main.go` via `parsers.Registry.Register()`.
- The registry tries versions in registration order; the first success wins.
- Add new bank formats as a new version directory, never by modifying an existing one.
- Parser output (`ParsedTransaction`) uses strings to preserve source representation;
  normalization to domain types happens in the service layer.

## Backend rules
- Keep HTTP handlers thin.
- Do not place business logic in handlers.
- Do not let repositories contain business rules.
- Prefer constructor-based dependency injection.
- Prefer interfaces only where they add real testability or boundary value.
- Keep package APIs small and explicit.
- Favor standard library first.

## Domain guidance
Core entities include:
- statements
- statement uploads
- transactions
- transaction identifiers (merchants)
- tags
- rules
- report aggregates

When adding behavior:
- preserve financial correctness
- treat money, dates, and identifiers carefully
- direction values are lowercase in the DB: `'debit'` and `'credit'`
- normalize transaction identifiers consistently
- avoid lossy transformations of source statement data

## Persistence guidance
- PostgreSQL is the source of truth for structured application data.
- Filesystem storage (`storage/statements/`) is used for raw PDF uploads.
- Schema changes must be added through migrations (DDL only — no SELECT in migrations).
- Avoid hidden schema assumptions in code.

## Migration rules
- Migration files contain only DDL: CREATE, ALTER, DROP, index creation.
- Never put SELECT, INSERT, or query-only statements in migration files.
- Reporting queries belong in `repository/report.go`.

## API design
- Use stable, predictable JSON shapes.
- Validate request payloads.
- Return clear error messages.
- Keep response contracts explicit.
- Avoid leaking internal persistence models directly if a transport DTO is clearer.

## Testing
Prefer:
- unit tests for services and parsing logic
- repository/integration tests where persistence behavior matters

When changing parsing or money logic, tests are strongly preferred.

## Code style
- Small functions
- Explicit names
- Minimal hidden side effects
- Avoid over-generic utility packages
- Keep errors wrapped with useful context

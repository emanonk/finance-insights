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
- parser layer → statement parsing and normalization

Preferred flow:
`handler -> service -> repository`

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
- transaction identifiers
- tags
- rules
- summaries

When adding behavior:
- preserve financial correctness
- treat money, dates, and identifiers carefully
- normalize transaction identifiers consistently
- avoid lossy transformations of source statement data

## Persistence guidance
- PostgreSQL is the source of truth for structured application data.
- Filesystem storage is used for raw uploads and generated outputs.
- Schema changes must be added through migrations.
- Avoid hidden schema assumptions in code.

## API design
- Use stable, predictable JSON shapes.
- Validate request payloads.
- Return clear error messages.
- Keep response contracts explicit.
- Avoid leaking internal persistence models directly if a transport DTO is clearer.

## Parsing guidance
- Parsers should be versioned and isolated by bank/source.
- New parser implementations should fit the existing registry/version pattern.
- Preserve the original file and raw reference data when useful for traceability.
- Prefer deterministic parsing over heuristic-heavy parsing unless explicitly required.

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
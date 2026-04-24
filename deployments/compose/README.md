# Docker Compose

Runs the full Finance Insights stack locally:

- `db`  — PostgreSQL 17
- `api` — Go backend (`apps/api`)
- `web` — React/Vite build served by nginx (`apps/web`)

The web container proxies `/api/*` to the `api` service, so the frontend
can call the backend using relative URLs (no CORS, no build-time URL).

## Quick start

From this directory:

```bash
cp .env.example .env
docker compose up --build
```

Then open:

- Web:  http://localhost:5173
- API:  http://localhost:8080/health
- DB:   `postgres://finance:finance@localhost:5432/finance_insights`

Stop with `Ctrl+C`, or `docker compose down` to remove containers.
Use `docker compose down -v` to also drop the database volume.

## Environment variables

Defined in `.env` (copied from `.env.example`):

| Variable            | Default             | Used by    |
| ------------------- | ------------------- | ---------- |
| `POSTGRES_USER`     | `finance`           | db, api    |
| `POSTGRES_PASSWORD` | `finance`           | db, api    |
| `POSTGRES_DB`       | `finance_insights`  | db, api    |
| `POSTGRES_PORT`     | `5432`              | db (host)  |
| `API_PORT`          | `8080`              | api (host) |
| `WEB_PORT`          | `5173`              | web (host) |

The API receives `DATABASE_URL` built from the Postgres variables.

## Storage

The repository `storage/` directory is mounted into the `api` container
at `/storage` for raw uploads and generated outputs.

## Notes

- Run commands from this directory so Compose picks up `.env`.
- Rebuild images after source changes with `docker compose up --build`.
- For day-to-day development with hot reload, prefer running `apps/api`
  and `apps/web` natively (see their READMEs); Compose is aimed at
  full-stack runs and reproducible environments.

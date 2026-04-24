# AGENTS.md

## Scope
This file applies to everything under `apps/web`.

## Tech
- node v22.15.0
- React
- TypeScript
- Vite
- Tailwind CSS v4

## Frontend architecture
Use feature-oriented organization where practical:
- `src/pages` for route-level pages
- `src/features/*` for domain-specific UI and API interactions
- `src/components/*` for shared presentation components
- `src/lib/*` for shared utilities and API client code

## Frontend rules
- Prefer function components and hooks.
- Use TypeScript strictly; avoid `any`.
- Keep components focused and composable.
- Keep data-fetching and API calls out of purely presentational components.
- Avoid global state unless there is a clear cross-page need.
- Prefer local state, lifted state, or feature-scoped hooks first.

## UI behavior
The product is a personal finance app.
Optimize for:
- clarity
- fast scanning
- low visual noise
- good table/filter workflows
- trustworthy presentation of money and dates

## Design guidance
- Use Tailwind utility classes directly unless repetition clearly justifies extraction.
- Keep layouts simple and readable.
- Prefer practical fintech-style UI over decorative design.
- Tables, filters, summaries, and detail drawers/modals should feel structured and predictable.

## Data handling
- Treat amounts, dates, currencies, and tags as first-class typed data.
- Format money and dates consistently.
- Do not bury transformation logic inside JSX when it can live in helpers/hooks.
- Keep frontend models clear when mapping from API responses.

## Routing and page design
Expected major page types include:
- dashboard
- statements
- transactions
- tag rules
- reports
- calendar

When building screens:
- start with usable structure
- then add polish
- keep empty states and loading states explicit

## Testing and quality
- Prefer simple testable logic in hooks/utilities.
- Avoid tightly coupling components to transport details.
- When adding reusable UI primitives, keep APIs small and obvious.

## What to avoid
- Do not introduce heavy state libraries early without a demonstrated need.
- Do not over-componentize tiny one-off markup.
- Do not create generic abstractions before repeated patterns exist.
- Do not mix mock/demo data into production flows.
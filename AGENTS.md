# AGENTS.md

## Project
Open-source personal expenses application.

Stack:
- Backend: Go 1.25
- Frontend: React + TypeScript + Vite + Tailwind CSS v4
- Database: PostgreSQL
- Infra: Docker Compose

## Repository structure
- `apps/api` → Go backend
- `apps/web` → React frontend
- `storage` → uploaded statements, processed outputs, exports
- `deployments/compose` → Docker Compose files

## Product concepts
- Statement: uploaded bank file
- Transaction: normalized financial record parsed from a statement
- Transaction Identifier: normalized merchant/source identifier reused across matching transactions
- Primary Tag: main category such as groceries, utilities, eating out
- Secondary Tags: optional extra classification labels
- Rule: logic that auto-applies metadata to transactions sharing the same identifier

## Global engineering rules
- Follow existing structure; do not invent new top-level folders without a strong reason.
- Prefer small, explicit, readable code over clever abstractions.
- Do not refactor unrelated files while implementing a focused task.
- Do not add new frameworks or libraries unless clearly justified.
- Keep changes minimal and local to the problem being solved.
- Update docs when behavior, setup, or architecture changes.

## Safety rails
- Never hardcode secrets, credentials, tokens, or local machine paths.
- Use environment variables for configuration.
- Preserve backward compatibility unless the task explicitly requires a breaking change.

## Delivery expectations
- For new features, include:
  - implementation
  - basic validation/error handling
  - tests where practical
  - brief docs or README updates when setup/usage changes

## Commands
- Backend dev: work inside `apps/api`
- Frontend dev: work inside `apps/web`
- Full stack: use Docker Compose from `deployments/compose`

## What to avoid
- Do not move files across layers without necessity.
- Do not mix UI concerns into backend code.
- Do not mix persistence concerns into React components.
- Do not introduce premature generic frameworks for parsing or tagging until needed.
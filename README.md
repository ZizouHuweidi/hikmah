# Sabeel

Sabeel is a social reading platform for discovering, tracking, annotating, and reviewing books.

The current foundation supports personal libraries, reading progress, notes, reviews, collections, and public profiles. V1 completes the books experience with a real catalog, social feeds, recommendations, responsive UX, and an installable PWA.

## Core Features

- Zitadel-backed registration and login with opaque Sabeel sessions.
- Book/source creation with metadata and contributors.
- Personal library tracking with status, progress, and visibility.
- Notes, reviews, and collections for organizing learning.
- Public profiles and public library views.

## Stack

- Backend: Go, Echo, PostgreSQL, pgx and sqlc.
- Frontend: React, React Router and Tailwind.
- Tooling: Podman Compose, Containerfiles, justfile commands.

## Local Development

Prerequisites: Go, Node/npm, Podman, Podman Compose, `just`.

```sh
cp .env.example .env
just up
just seed
```

Open the frontend at `http://localhost:3000`.

Useful URLs:

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8080`
- Health: `http://localhost:8080/health`
- Readiness: `http://localhost:8080/ready`
- Identity: `http://127.0.0.1:8081`
- Verification email: `http://127.0.0.1:8025`

## Common Commands

- `just up` - Start local services.
- `just down` - Stop local services.
- `just migrate` - Apply database migrations.
- `just migrate-status` - Show migration status.
- `just migrate-create name` - Create a migration file.
- `just seed` - Seed demo data.
- `just test` - Run Go tests.
- `just check` - Run sqlc checks, Go tests, and frontend checks.
- `just frontend-dev` - Start the frontend dev server.
- `just frontend-build` - Build the frontend.
- `just prod-config` - Validate the independent production stack configuration.
- `just prod-build` - Build the production application images once.
- `just prod-up` - Start Sabeel and its dedicated production identity stack.

## Database Workflow

- Schema changes live in `migrations/` and are applied with Goose through `cmd/migrate`.
- Application queries live in `internal/db/queries/` and generate Go code with sqlc.

## API Exploration

Open the `bruno/` directory in Bruno and select the `local` environment. The collection covers auth, profile, sources, books, library, notes, reviews, and collections.

See [`docs/v1-plan.md`](docs/v1-plan.md) for the V1 scope, current status, delivery sequence, and release criteria.

See [`docs/identity.md`](docs/identity.md) for the Zitadel client, browser flow, migration behavior, and production requirements.

See [`docs/deployment.md`](docs/deployment.md) for production deployment, backup, and recovery.

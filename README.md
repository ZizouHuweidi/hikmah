# Sabeel

Sabeel is a personal knowledge library for collecting, tracking, annotating, and reviewing sources such as books and other learning media.

The MVP focuses on helping a reader build a library, track reading progress, write notes and reviews, organize sources into collections, and publish a lightweight public profile.

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
just migrate
just seed
just frontend-dev
```

Open the frontend at `http://localhost:3000`.

Useful URLs:

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8080`
- Health: `http://localhost:8080/health`
- Readiness: `http://localhost:8080/ready`
- Identity: `http://127.0.0.1:8081`

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

## Database Workflow

- Schema changes live in `migrations/` and are applied with Goose through `cmd/migrate`.
- Application queries live in `internal/db/queries/` and generate Go code with sqlc.

## API Exploration

Open the `bruno/` directory in Bruno and select the `local` environment. The collection covers auth, profile, sources, books, library, notes, reviews, and collections.

See [`docs/identity.md`](docs/identity.md) for the Zitadel client, browser flow, migration behavior, and production requirements.

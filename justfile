set dotenv-load

compose := "podman compose"
database_url := env_var_or_default("DATABASE_URL", "postgres://sabeel:sabeel@localhost:5432/sabeel?sslmode=disable")
production_env := env_var_or_default("SABEEL_PRODUCTION_ENV", ".env.production")
sqlc_image := "docker.io/sqlc/sqlc:1.30.0"

default:
    just --list

up: compose-build
    {{ compose }} up -d

down:
    {{ compose }} down

logs:
    {{ compose }} logs -f

ps:
    {{ compose }} ps

prod-config:
    {{ compose }} --env-file {{ production_env }} -f compose.production.yaml config --quiet

prod-build: prod-config
    {{ compose }} --env-file {{ production_env }} -f compose.production.yaml build app frontend

prod-up: prod-build
    {{ compose }} --env-file {{ production_env }} -f compose.production.yaml up -d

prod-down:
    {{ compose }} --env-file {{ production_env }} -f compose.production.yaml down

prod-logs:
    {{ compose }} --env-file {{ production_env }} -f compose.production.yaml logs -f

prod-ps:
    {{ compose }} --env-file {{ production_env }} -f compose.production.yaml ps

build:
    just backend-image

images:
    just backend-image
    just frontend-image

backend-image:
    podman build -f Containerfile -t sabeel:latest .

frontend-image:
    podman build -f frontend/Containerfile -t sabeel-frontend:latest frontend

compose-build:
    {{ compose }} build

run:
    go run ./cmd/server

test:
    go test ./...

check:
	just db-check
	just test
	just frontend-check

backend-check:
	just db-check
	just test

fmt:
    gofmt -w cmd internal pkg

tidy:
    go mod tidy

sqlc-generate:
    podman run --rm -v "$PWD:/src:Z" -w /src {{ sqlc_image }} generate

sqlc-vet:
    podman run --rm -v "$PWD:/src:Z" -w /src {{ sqlc_image }} vet

db-generate: sqlc-generate

db-check:
    just sqlc-generate
    just sqlc-vet

migrate:
    DATABASE_URL='{{ database_url }}' go run ./cmd/migrate up

migrate-down:
    DATABASE_URL='{{ database_url }}' go run ./cmd/migrate down

migrate-status:
    DATABASE_URL='{{ database_url }}' go run ./cmd/migrate status

migrate-create name:
    go run ./cmd/migrate create {{ name }}

seed:
    DATABASE_URL='{{ database_url }}' go run ./cmd/seed

db-shell:
    {{ compose }} exec postgres psql -U sabeel -d sabeel

health:
    curl -fsS http://localhost:8080/health

ready:
    curl -fsS http://localhost:8080/ready

frontend-health:
    curl -fsS http://localhost:3000/

frontend-install:
    npm --prefix frontend install

frontend-dev:
    npm --prefix frontend run dev

frontend-start:
    npm --prefix frontend run start

frontend-build:
    npm --prefix frontend run build

frontend-check:
    npm --prefix frontend run check

frontend-typecheck:
    npm --prefix frontend run typecheck

frontend-lint:
    npm --prefix frontend run lint

frontend-format:
    npm --prefix frontend run format

frontend-ci:
    just frontend-check
    just frontend-typecheck
    just frontend-build

dev: up

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Server-rendered to-do list web app built with Go + [Buffalo](https://gobuffalo.io) (v1.1) and Pop/PostgreSQL. Go module name is `todo`. No JS build pipeline: plain CSS/JS in `public/assets/`, HTML via Plush templates.

## Commands

PostgreSQL must be running first (`docker compose up -d db`; the dev container does this on start). `DB_HOST` overrides the default `127.0.0.1`.

```bash
buffalo dev                          # dev server with live reload on :3000 (builds ./cmd/app, see .buffalo.dev.yml)
go run ./cmd/app                     # run without live reload
docker compose up --build            # app + db in containers

GO_ENV=test soda migrate -e test     # prepare the test DB (soda: go install github.com/gobuffalo/pop/v6/soda@v6.4.1)
GO_ENV=test go test -p 1 ./...       # all tests; -p 1 is required (packages share one DB)
GO_ENV=test go test -p 1 ./actions -run Test_ActionSuite -testify.m Test_TasksCreate_Valid   # single test

gofmt -l .                           # CI fails if this lists any file
go vet ./...
```

CI (`.github/workflows/ci.yaml`) runs gofmt check, vet, migrations, tests and build, plus a Docker image build.

## Architecture

- **Entry point** `cmd/app/main.go`: runs `models.Migrate()` (applies embedded migrations) then serves `actions.App()`. Migrations therefore run automatically on every app start, including in Docker/production.
- **DB connection** is opened in `models/models.go`'s `init()` using `GO_ENV` (default `development`) against `database.yml`. Importing `models` connects to the DB, so any test or binary that imports it needs a reachable Postgres. Production uses `DATABASE_URL`.
- **Everything is embedded** via `embed.FS`: `migrations/`, `templates/`, `public/` each expose `FS()`. New files in those directories are picked up automatically, but templates in deeper subdirectories need the `templates/embed.go` glob (`* */*`) extended.
- **Routing/middleware** in `actions/app.go`: force-SSL (production only), CSRF, and `popmw.Transaction` which wraps every request in a transaction — handlers get it with `c.Value("tx").(*pop.Connection)`. Routes are registered explicitly (not `app.Resource`).
- **Single-page UI**: all task views go through `renderIndex` in `actions/tasks.go`, which loads all tasks and renders `tasks/index.plush.html` with an `indexView` (create form state + inline edit state). Edit (`GET /tasks/{id}/edit`) re-renders the index with one row in edit mode rather than a separate page. Validation failures re-render the index with status 422; successes redirect with 303 + flash.
- **Forms** use POST with a hidden `_method` (PUT/DELETE) and `authenticity_token` for CSRF. `public/assets/app.js` adds a confirm dialog for forms with `data-confirm`.
- **Mass-assignment guard**: `models.Task` fields other than `Title` are tagged `form:"-"`, so `c.Bind` only accepts the title. Keep this when adding fields. Handlers trim the title; `Task.Validate` enforces non-empty and `TitleMaxLength` (200 runes, matching the column size).
- **Migrations**: schema in `.fizz`, data seed in `.sql` (`20260901000001_seed_tasks` inserts fixed-UUID demo tasks idempotently).

## Tests

Test suites use `gobuffalo/suite/v4` (`ActionSuite` in `actions/actions_test.go`, `ModelSuite` in `models/models_test.go`); individual tests are methods `func (as *ActionSuite) Test_...()`. The suite truncates tables between tests. Fixtures live in `fixtures/*.toml` and are loaded by scenario name, e.g. `as.LoadFixture("tres tareas")`.

## Conventions

- User-facing text, comments, test fixture data and commit messages are in **Spanish**.
- Commit messages follow Conventional Commits with Spanish descriptions (`feat: ...`, `build: ...`, `ci: ...`).

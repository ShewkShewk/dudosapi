# Development

Local setup, migrations, code generation, and project layout for working on
dudosapi. See [`README.md`](../README.md) for what the service does and its
API, [`CLAUDE.md`](../CLAUDE.md) for the auth model and deployment, and
[`TESTING.md`](./TESTING.md) for the end-to-end test suite.

## Prerequisites

- Go (see `go.mod` for the version)
- Docker, for a local Postgres instance
- The `migrate` CLI: `brew install golang-migrate` (or `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)
- The `sqlc` CLI: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- The `templ` CLI: `go install github.com/a-h/templ/cmd/templ@latest`

## Local development setup

1. **Start Postgres:**

   ```sh
   docker-compose up -d
   ```

2. **Create a `.env` file** in the repo root (gitignored, never commit it) with:

   ```sh
   TABROOM_HOSTNAME=
   TABROOM_USERNAME=
   TABROOM_PASSWORD=
   DB_CONNECTION_STRING=postgres://myuser:mypassword@localhost:5432/dudos_duda
   LOCAL_DB_URL=postgres://myuser:mypassword@localhost:5432/dudos_duda?sslmode=disable
   ```

   `TABROOM_*` are credentials for the Tabroom API (`tbapi`). `DB_CONNECTION_STRING`
   is what the app itself connects with; `LOCAL_DB_URL` is a separate var read by
   `migrate_db.sh` (it takes a `sslmode` query param the app doesn't need).
   `PORT` is optional and defaults to `8080`.

3. **Run migrations:**

   ```sh
   ./migrate_db.sh local up
   ```

4. **Run the service:**

   ```sh
   go run .
   ```

   It listens on `:8080` (or `$PORT`) with no auth locally — see `CLAUDE.md` for
   why that's fine on localhost but not elsewhere.

`requests.http` has example requests for every endpoint (works with the
JetBrains HTTP client or the VS Code REST Client extension).

## Database migrations

`migrate_db.sh <local|prod> <migrate args...>` wraps the `migrate` CLI, sourcing
`LOCAL_DB_URL` / `PROD_DB_URL` from `.env`.

```sh
./migrate_db.sh local up          # apply all pending migrations
./migrate_db.sh local down 1      # roll back one migration
./migrate_db.sh prod up           # against production — be careful
```

New migrations go in `db/migrations/`, following the existing
`<n>_<description>.up.sql` / `.down.sql` naming.

## Code generation

Two things are generated and checked in — regenerate them after editing their
sources, don't hand-edit the output:

- **`internal/db/sqlc/`** — generated from `schema.sql` + `query.sql` by `sqlc generate`.
  `schema.sql` is maintained separately from `db/migrations/` and needs to be kept
  in sync by hand when a migration changes the schema.
- **`pairings_templ.go`, `schools_templ.go`** — generated from `pairings.templ` /
  `schools.templ` by `templ generate`.

## Project layout

```
main.go              entry point, config from env
routes.go             HTTP handlers, GCS publishing
import.go             Tabroom -> Postgres import logic
tournaments.go         pairings/school-status read logic
domain.go              API response types
encode.go               JSON encode/decode helpers
summary.go              /summary handler logic
pairings.templ, schools.templ   published HTML pages
db/migrations/           schema migrations (source of truth for the DB schema)
schema.sql, query.sql     sqlc input (schema.sql must be kept in sync with migrations)
internal/db/sqlc/          sqlc-generated Go (don't hand-edit)
```

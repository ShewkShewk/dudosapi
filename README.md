# dudosapi

Go service that ingests debate tournament data from Tabroom via
[`tbapi`](https://github.com/ShewkShewk/tbapi), normalizes it into Postgres, and
publishes two rendered HTML pages — pairings and school check-in status — for
tournament-day displays to poll.

## Documentation

- [`docs/DEVELOPMENT.md`](./docs/DEVELOPMENT.md) — local setup, migrations, code generation, project layout
- [`docs/TESTING.md`](./docs/TESTING.md) — the end-to-end test suite
- [`CLAUDE.md`](./CLAUDE.md) — auth model (Cloud Run IAM) and deployment notes

## Tech stack

- Go, standard library `net/http` (no framework)
- Postgres, queried through [`sqlc`](https://sqlc.dev) (`internal/db/sqlc`, generated — don't hand-edit)
- [`golang-migrate`](https://github.com/golang-migrate/migrate) for schema migrations (`db/migrations/`)
- [`templ`](https://templ.guide) for the two published HTML pages (`pairings.templ`, `schools.templ`)

## API

| Method | Path                                 | Description                                              |
|--------|---------------------------------------|------------------------------------------------------------|
| GET    | `/tournaments`                        | List tournaments from Tabroom, annotated with local import status |
| POST   | `/tournaments/{id}/import`            | Import a tournament from Tabroom and publish pairings/status |
| DELETE | `/tournaments/{id}`                   | Delete an imported tournament                             |
| GET    | `/tournaments/{id}/pairings/latest`   | Latest pairings for a tournament, as JSON                 |
| GET    | `/tournaments/{id}/schools/status`    | School check-in status for a tournament, as JSON          |
| GET    | `/summary`                            | Tournament and completed-round counts                     |

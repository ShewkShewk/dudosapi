# Testing

There's currently one layer of automated testing: an end-to-end suite that
stands up every local dependency the app needs (Tabroom, Postgres, GCS) and
drives the real, unmodified HTTP handlers through a full import.

## Running it

```sh
go test -tags=e2e ./...
```

Requires Docker running locally (Postgres and the GCS emulator run as
containers via [testcontainers-go](https://golang.testcontainers.org/)).
Nothing else needs to be running or configured first — the test binary
brings up and tears down everything it needs.

The tests are gated behind the `e2e` build tag specifically so `go test
./...` (no tag) stays fast and Docker-free for routine use; only the tagged
run needs Docker.

Run a single test while iterating:

```sh
go test -tags=e2e -run TestImportScenarios -v ./...
```

## What it stands up, and why

Real Tabroom, Postgres, and GCS aren't used — each is replaced with
something that speaks the same protocol, so the real `NewServer(config)`
handler runs completely unmodified against them:

- **Tabroom** — a plain `httptest.Server` (`e2e_tabroom_fake_test.go`)
  implementing the four endpoints `tbapi.TabroomApi` actually calls: the
  login page (scraped for a salt/sha form), the login POST (sets a
  `TabroomToken` cookie), the tournament listing (scraped HTML), and
  `download_data.mhtml` (returns a fixture's JSON verbatim). No real Tabroom
  account is needed, and results are deterministic.
- **Postgres** — a real `postgres:16-alpine` container
  (`e2e_postgres_test.go`), started once per test binary run in `TestMain`.
  `db/migrations/` is applied through the `golang-migrate` Go library
  (not the `migrate` CLI), then the freshly-migrated database is snapshotted.
  Every test calls `resetDB(t)`, which restores that snapshot in `t.Cleanup`
  — each test gets a database that's freshly migrated and empty, regardless
  of what any other test did.
- **GCS** — `fsouza/fake-gcs-server` (`e2e_gcs_test.go`). The Go
  `cloud.google.com/go/storage` client honors `STORAGE_EMULATOR_HOST`
  natively (including disabling auth), so `getStorageClient` in `routes.go`
  needs no test-specific code at all. The one non-obvious part: fake-gcs-server
  needs to know its own externally-reachable address (`FAKE_GCS_EXTERNAL_URL`
  / `FAKE_GCS_PUBLIC_HOST` env vars) *at container startup*, or uploads
  silently succeed while reads 404 (the client's Reader hits an XML-style
  route the emulator only serves for its configured public host). Since
  those need to be set before the container starts, the test reserves a free
  host port itself first and requests a fixed port binding, rather than
  using testcontainers' usual dynamic port allocation.

## File layout

| File | Contents |
|------|----------|
| `e2e_test.go` | `TestImportScenarios` — the table-driven entrypoint |
| `e2e_scenarios_test.go` | `e2eScenario` type + scenario builders (e.g. `goldenPathScenario`) |
| `e2e_driver_test.go` | `runScenario` (drives one scenario through the real handler) + `assertDeepEqual` |
| `e2e_tabroom_fake_test.go` | Fake Tabroom HTTP server |
| `e2e_tabroom_fake_smoke_test.go` | Proves the fake server + fixture work against a real `tbapi` client |
| `e2e_postgres_test.go` | `TestMain`, Postgres container + migrations + snapshot, `resetDB` |
| `e2e_postgres_smoke_test.go` | Proves migrations ran and `resetDB` actually isolates tests |
| `e2e_gcs_test.go` | GCS emulator container setup |
| `e2e_gcs_smoke_test.go` | Proves the real `getStorageClient`/`uploadComponent` round-trip against it |
| `e2e_http_helpers_test.go` | `postImport`, `getJSON` |
| `e2e_gcs_helpers_test.go` | `readGcsBlob` |
| `testdata/e2e/*.json` | Fixture Tabroom tournament payloads |

## Adding a new scenario

1. Add a fixture JSON under `testdata/e2e/` — a `tbapi.TournamentData`
   payload, trimmed to whatever fields `import.go` actually reads (it
   doesn't need to be a complete Tabroom export; `encoding/json` doesn't
   require unrecognized/missing fields).
2. Add a builder function in `e2e_scenarios_test.go` (following
   `goldenPathScenario`) returning an `e2eScenario` with the fixture path,
   tournament id/date/name, and the exact expected `TournamentPairings`,
   `TournamentSchoolsStatus`, and `Summary` values.
3. Append it to the `scenarios` slice in `TestImportScenarios`
   (`e2e_test.go`).

No new assertion code is needed — `runScenario` and `assertDeepEqual` are
shared by every scenario; only the expected values change.

### A note on expected values

`assertDeepEqual` does a full `reflect.DeepEqual` against the expected
struct, not a spot-check of a few fields — so expected values need to be
exact, including ones easy to get wrong:

- **Times are converted to `America/Chicago`** (hardcoded in
  `getTimezone()`), regardless of what timezone the fixture's timestamps are
  written in — Go's `time.Parse` with no zone in the layout treats the
  fixture's wall-clock digits as UTC, so e.g. `"09:00:00"` in a fixture
  becomes `4:00AM` in an August (daylight-saving) expected value.
- **`Flighted` is `true` only when a pairing's flight number is `> 1`** —
  a single-flight round is `Flighted: false` even though it has one flight.

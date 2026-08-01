# dudosapi

Go service that ingests debate tournament data from Tabroom via
`github.com/ShewkShewk/tbapi`, normalizes it into Postgres (sqlc), and publishes
two rendered HTML pages — pairings and school check-in status — to a GCS bucket
that tournament-day displays poll.

Deployed on Cloud Run. The service URL is deliberately not recorded here, because
this repository is public — get it from the GCP console or `gcloud run services
list`.

## Auth is enforced by Cloud Run IAM, not by this code

**There is deliberately no auth middleware in `routes.go`.** Every handler is
registered unguarded, but the deployed service is not open: Cloud Run rejects
unauthenticated requests at the edge, before they reach the Go process.

Verified 2026-08-01 — an unauthenticated `GET /summary` against the deployed
service returns `403 Forbidden` as an HTML page from Google's front end. (This
app has no 403 path of its own, so an HTML 403 is always the platform, never the
application.)

What this means when working on this code:

- **Do not "fix" the missing auth** by adding a shared bearer token or API key.
  IAM already gives per-identity access, revocation without a redeploy, and
  Cloud Audit Logs. A static secret in `.env` would be strictly less capable and
  one more thing to leak.
- Access is granted through the Cloud Run Invoker IAM role in GCP, not in this
  repo. Nothing here changes who can call the API.
- New endpoints are protected automatically. There is nothing to wire up.
- If this service is ever run **outside** Cloud Run — plain Docker, a VM, or
  `go run` reachable beyond localhost — it has no authentication whatsoever.
  That is the one scenario where app-level auth would need to be added.

Calling the deployed API by hand:

```sh
curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" "$HOST/summary"
```

Identity tokens last one hour, so a stale token shows up as a 403 — refresh it
rather than debugging the app. Note that `gcloud auth print-identity-token` mints
a token whose audience is
the gcloud client ID rather than the service URL; if a *freshly minted* token
still returns 403, pass `--audiences=<service URL>`.

## Published output is public by design

`publishPairings` and `publishSchoolsStatus` in `routes.go` upload to the
`duda_pairings` bucket at the end of every import. Both objects are **publicly
readable with no auth** (verified 2026-08-01). That is intentional: displays and
the public site fetch them directly, and pairings are public information at a
tournament.

So the IAM gate protects *mutations*, not the data. Anything rendered into
`pairings.templ` or `schools.templ` becomes world-readable the moment an import
runs. Do not add anything to those templates that should not be public — judge
emails and student names (as opposed to entry codes) are the obvious hazards, and
both are already present in the database.

Objects are written with `CacheControl: no-store`, and both templates carry
`<meta http-equiv="refresh" content="60">`, so a display picks up a new import
within a minute without cache busting.

Object names are fixed — `pairings.html` and `status.html` — regardless of
tournament id, so only one tournament's output is live at a time. Importing a
second tournament overwrites the first one's published pages.

## Local development

`.env` supplies config through `godotenv/autoload` (blank-imported in `main.go`),
and is gitignored. Local runs have no auth at all, which is fine against
localhost.

`migrate_db.sh <local|prod> <migrate args...>` reads `LOCAL_DB_URL` /
`PROD_DB_URL`, sourcing `.env` when present and otherwise using the ambient
environment.

This repository is **public**. Never commit `.env`, the compiled `dudosapi`
binary, or any untracked file holding credentials, tokens, or connection strings
— several exist locally and none of them belong in git. Check `git status` before
staging, and prefer `git add <path>` over `git add .`.

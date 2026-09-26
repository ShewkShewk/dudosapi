---
name: "e2e-fixture"
description: "Generate a new e2e test scenario for dudosapi from a real Tabroom tournament JSON export — trims it to one edge case, strips judge/student PII, and derives expected values by actually running the import instead of hand-computing them. Use when the user wants a new e2e test case built from real tournament data, or asks to add/generate a fixture/scenario under testdata/e2e/."
---

Automates the "Adding a new scenario" recipe in `docs/TESTING.md` and
`CLAUDE.md`, with two things manual authoring tends to get wrong: real PII
ending up in a public repo, and hand-computed expected values drifting from
what the code actually does (the timezone and `Flighted` traps both docs
call out).

Invocation: `/e2e-fixture <path-to-real-tournament-json> <description of the
scenario/edge case to capture>`, e.g. `/e2e-fixture
~/tourneys/state2026.json a round with a bye`. If either argument is
missing, ask for it rather than guessing.

## 1. Inspect the source

Don't read the whole file into context if it's large — use `jq`/`python3
-m json.tool` to check sizes first (`categories | length`, judge/school/
student counts, round counts per event). The shape is
`github.com/ShewkShewk/tbapi`'s `TournamentData`; run `go doc
github.com/ShewkShewk/tbapi TournamentData` if you need the field layout.

## 2. Trim to the requested edge case

Keep only the categories/events/schools/sites/judges/students/rounds needed
to exercise the described scenario — existing fixtures
(`testdata/e2e/fixture_*.json`) run 150-400 lines; that's the ceiling to aim
under, not a floor. Preserve referential integrity: ids linking
judges/schools/students/entries/sections must stay consistent after
deleting the rest.

## 3. Anonymize before anything touches disk

This repo is public (`CLAUDE.md`). Real judge/student data must never land
in a commit, including in fixture JSON that never gets rendered anywhere.

- **Judges** (`First`, `Last`, `Email`, `Phone`) and **students** (`First`,
  `Last`) are real PII in a raw Tabroom export — replace every one with a
  placeholder, reusing the existing fixtures' convention (e.g. judge "Jane
  Judge" / `jane.judge@example.com`, students "Alice Anderson" / "Bob
  Baker" paired with entry codes like `AA1`/`BB1`). Keep one consistent
  real→fake mapping per unique person for the duration of the trim so
  relationships in the data stay coherent — but keep that mapping only in
  memory or in the session scratchpad, never write it into the repo.
- **School names** — default to anonymizing these too (e.g. "Alpha High" /
  "Beta High"), since the point of this exercise is data that's safe to
  commit. If the user explicitly wants realistic school names kept, that's
  fine, but don't do it by default.
- **Entry codes** (e.g. `AA1`) aren't PII themselves — regenerate them from
  the placeholder names if needed, matching the existing style.
- **Judge ids/person ids, student ids, and school `chapter` ids** are real
  Tabroom record identifiers tied to real people and schools — regenerate
  these too rather than carrying the source file's numbers over, even though
  they're not human-readable PII by themselves. Pick fresh placeholder
  numbers (small sequential values are fine, e.g. `90001`, `90002`) and
  rewrite every reference consistently: a judge's `id` appears again in each
  ballot's `judge` field, a student's `id` appears again in their entry's
  `students` list, and a school's `chapter` appears again in that school's
  own `students[].chapter`. Entry ids, round ids, section ids, ballot ids,
  and site/room ids aren't tied to a person or school, so those can stay as
  in the source.
- Skim the trimmed file yourself for anything that still looks like a real
  name, email, phone number, or an untouched source id in one of the fields
  above before writing it — don't rely on having caught every field on the
  first pass.

## 4. Pick a fixture path and tournament id

`grep -n 'tournamentID:' e2e_scenarios_test.go` to see ids already in use
(currently the `990xx` range) and pick an unused one. Name the fixture
`testdata/e2e/fixture_<slug>.json` where slug describes the scenario.

## 5. Scaffold the scenario with placeholder expectations

Add a builder function to `e2e_scenarios_test.go` following
`goldenPathScenario`'s shape — fixture path, tournament id/date/name filled
in, but `wantPairings`, `wantSchoolsStatus`, and `wantSummary` left as
zero-value structs. These are placeholders you'll overwrite in step 7 with
real output, not a guess to get right now. Append `solo(yourNewScenario())`
to the `sequences` slice in `TestImportScenarios` (`e2e_test.go`).

## 6. Run it to get real output

```sh
go test -tags=e2e -run TestImportScenarios -v ./...
```

Requires Docker running — if it's not available, stop and tell the user
rather than guessing at expected values by hand. The zero-valued want
fields will fail `assertDeepEqual` and print the actual `--- got ---` JSON
for pairings and schools status; the summary mismatch prints separately.
This is the authoritative expected output — deriving it this way is the
whole reason to prefer this workflow over hand-computing, since it can't
diverge from what `getTimezone()` and the `Flighted` logic actually produce.

## 7. Fill in the real expected values

Convert the got JSON into Go struct literals matching the existing style
(`TournamentPairings{...}`, `ptr(...)` for optional fields, `WIN`/`LOSS`
constants) and paste them into the builder in place of the zero values. Do
a sanity pass on the two documented traps rather than re-deriving them:
times should already read as `America/Chicago` clock time, and `Flighted`
should be `true` only where a pairing's flight number is `> 1`. Also fill
in `wantPairingsHTMLContains`/`wantStatusHTMLContains` with a few
distinguishing substrings (entry names, room, judge name, the checked-in
color spans), following the existing scenarios' style.

## 8. Confirm it passes clean

Re-run `go test -tags=e2e -run TestImportScenarios -v ./...`. It should
pass with no diff now that the want fields are real.

## 9. Report back

Summarize what was added: fixture path, scenario name, tournament id, and
which files changed (`testdata/e2e/fixture_*.json`,
`e2e_scenarios_test.go`, `e2e_test.go`). Remind the user the fixture holds
anonymized-but-structurally-real tournament data and is about to be
committed to a public repo — worth a skim before they push if they want to
double-check the anonymization themselves.

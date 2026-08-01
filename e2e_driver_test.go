//go:build e2e

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

// e2eScenario describes one full import-and-verify pass: which fixture the
// fake Tabroom server should serve, and what the app is expected to produce
// for it. New scenarios (byes, forfeits, re-import, ...) are added to
// importScenarios below without touching runScenario or the comparison
// logic - only the expected values change.
type e2eScenario struct {
	name           string
	tournamentID   int
	tournamentDate string // YYYY-MM-DD, matches the fake Tabroom listing's date column
	tournamentName string
	fixturePath    string

	wantPairings      TournamentPairings
	wantSchoolsStatus TournamentSchoolsStatus
	wantSummary       Summary

	wantPairingsHTMLContains []string
	wantStatusHTMLContains   []string
}

func TestImportScenarios(t *testing.T) {
	scenarios := []e2eScenario{
		goldenPathScenario(),
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			runScenario(t, sc)
		})
	}
}

// goldenPathScenario: one tournament, two schools, one flighted round with
// a single decided ballot and a speaker award.
func goldenPathScenario() e2eScenario {
	return e2eScenario{
		name:           "golden path",
		tournamentID:   99001,
		tournamentDate: "2026-08-01",
		tournamentName: "Fixture Debate Invitational",
		fixturePath:    "testdata/e2e/fixture_tournament.json",

		wantPairings: TournamentPairings{
			Name:       "Fixture Debate Invitational",
			UpdateTime: "2026-08-01 1:05PM",
			EventPairings: []EventPairing{
				{
					Name:      "Public Forum",
					Number:    1,
					Flighted:  false,
					StartTime: "4:00AM",
					Pairings: []Pairing{
						{
							SectionId: 1,
							Flight:    1,
							Room:      ptr("Room 101"),
							AffEntry:  &Entry{Id: 10, Name: "AA1"},
							AffResult: ptr(WIN),
							NegEntry:  &Entry{Id: 20, Name: "BB1"},
							NegResult: ptr(LOSS),
							Judges: []Judge{
								{Id: 1, PersonId: 501, Name: "Jane Judge", Started: true},
							},
						},
					},
				},
			},
		},

		wantSchoolsStatus: TournamentSchoolsStatus{
			Name:       "Fixture Debate Invitational",
			UpdateTime: "2026-08-01 1:05PM",
			SchoolsStatus: []SchoolStatus{
				{Id: 1, Name: "Alpha High", CheckedIn: true},
				{Id: 2, Name: "Beta High", CheckedIn: false},
			},
		},

		wantSummary: Summary{TournamentCount: 1, RoundCount: 1},

		wantPairingsHTMLContains: []string{"AA1", "BB1", "Room 101", "Jane Judge", "Public Forum Round #1"},
		wantStatusHTMLContains:   []string{"Alpha High", "Beta High"},
	}
}

// runScenario drives the real, unmodified NewServer handler through the
// whole workflow - import a tournament from a fake Tabroom server built
// from the scenario's fixture, then read it back via every read endpoint -
// and checks both the JSON API responses and what actually landed in the
// published GCS bucket against the scenario's expected values.
func runScenario(t *testing.T, sc e2eScenario) {
	t.Helper()
	resetDB(t)
	ctx := context.Background()

	fixture, err := os.ReadFile(sc.fixturePath)
	if err != nil {
		t.Fatalf("read fixture %s: %v", sc.fixturePath, err)
	}
	tabroomServer := newFakeTabroomServer(fakeTabroomTournament{
		id:       sc.tournamentID,
		date:     sc.tournamentDate,
		name:     sc.tournamentName,
		dataJSON: fixture,
	})
	defer tabroomServer.Close()

	cfg := &Config{
		tabroomConfig: &TabroomConfig{
			hostname: tabroomServer.URL,
			username: "test-user",
			password: "test-password",
		},
		dbConnectionString: pgDSN(ctx),
	}

	handler, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	appServer := httptest.NewServer(handler)
	defer appServer.Close()

	postImport(t, fmt.Sprintf("%s/tournaments/%d/import", appServer.URL, sc.tournamentID))

	var pairings TournamentPairings
	getJSON(t, fmt.Sprintf("%s/tournaments/%d/pairings/latest", appServer.URL, sc.tournamentID), &pairings)
	assertDeepEqual(t, "pairings", sc.wantPairings, pairings)

	var status TournamentSchoolsStatus
	getJSON(t, fmt.Sprintf("%s/tournaments/%d/schools/status", appServer.URL, sc.tournamentID), &status)
	assertDeepEqual(t, "schools status", sc.wantSchoolsStatus, status)

	var summary Summary
	getJSON(t, appServer.URL+"/summary", &summary)
	assertDeepEqual(t, "summary", sc.wantSummary, summary)

	storageClient, err := getStorageClient(ctx)
	if err != nil {
		t.Fatalf("getStorageClient: %v", err)
	}
	defer storageClient.Close()

	pairingsHTML := readGcsBlob(t, ctx, storageClient, "pairings.html")
	for _, want := range sc.wantPairingsHTMLContains {
		if !strings.Contains(pairingsHTML, want) {
			t.Errorf("pairings.html missing %q", want)
		}
	}

	statusHTML := readGcsBlob(t, ctx, storageClient, "status.html")
	for _, want := range sc.wantStatusHTMLContains {
		if !strings.Contains(statusHTML, want) {
			t.Errorf("status.html missing %q", want)
		}
	}
}

// assertDeepEqual compares got against want and, on mismatch, prints both
// as indented JSON so it's clear which field(s) differ - one comparison
// function shared by every scenario instead of a bespoke assert function
// per test.
func assertDeepEqual(t *testing.T, label string, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		wantJSON, _ := json.MarshalIndent(want, "", "  ")
		gotJSON, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("%s mismatch:\n--- want ---\n%s\n--- got ---\n%s", label, wantJSON, gotJSON)
	}
}

func ptr[T any](v T) *T {
	return &v
}

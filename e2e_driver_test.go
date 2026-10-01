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

// runScenario drives the real, unmodified NewServer handler through the
// whole workflow - import one or more tournaments from a fake Tabroom server
// seeded with every scenario's fixture, then read each one back via every
// read endpoint - and checks both the JSON API responses and what actually
// landed in the published GCS bucket against each scenario's expected
// values.
//
// Scenarios are imported in the order given, and pairings/schools-status/HTML
// are asserted immediately after each import - so passing more than one
// scenario also exercises (and proves) that a later import overwrites the
// previously published pairings.html/status.html. Summary counts are global
// across all tournaments in the DB, so only the last scenario's wantSummary
// is checked, once, after every import has run; earlier scenarios' wantSummary
// fields are ignored when composing a multi-tournament run.
func runScenario(t *testing.T, scenarios ...e2eScenario) {
	t.Helper()
	if len(scenarios) == 0 {
		t.Fatal("runScenario: no scenarios given")
	}
	resetDB(t)
	ctx := context.Background()

	tabroomTournaments := make([]fakeTabroomTournament, len(scenarios))
	for i, sc := range scenarios {
		fixture, err := os.ReadFile(sc.fixturePath)
		if err != nil {
			t.Fatalf("read fixture %s: %v", sc.fixturePath, err)
		}
		tabroomTournaments[i] = fakeTabroomTournament{
			id:       sc.tournamentID,
			date:     sc.tournamentDate,
			name:     sc.tournamentName,
			dataJSON: fixture,
		}
	}
	tabroomServer := newFakeTabroomServer(tabroomTournaments...)
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

	storageClient, err := getStorageClient(ctx)
	if err != nil {
		t.Fatalf("getStorageClient: %v", err)
	}
	defer storageClient.Close()

	for _, sc := range scenarios {
		postImport(t, fmt.Sprintf("%s/tournaments/%d/import", appServer.URL, sc.tournamentID))

		var pairings TournamentPairings
		getJSON(t, fmt.Sprintf("%s/tournaments/%d/pairings/latest", appServer.URL, sc.tournamentID), &pairings)
		assertDeepEqual(t, "pairings", sc.wantPairings, pairings)

		var status TournamentSchoolsStatus
		getJSON(t, fmt.Sprintf("%s/tournaments/%d/schools/status", appServer.URL, sc.tournamentID), &status)
		assertDeepEqual(t, "schools status", sc.wantSchoolsStatus, status)

		var eventSchoolCounts TournamentEventSchoolCounts
		getJSON(t, fmt.Sprintf("%s/tournaments/%d/events/schools", appServer.URL, sc.tournamentID), &eventSchoolCounts)
		assertDeepEqual(t, "event school counts", sc.wantEventSchoolCounts, eventSchoolCounts)

		var schools = make([]School, len(sc.wantSchools))
		getJSON(t, fmt.Sprintf("%s/schools", appServer.URL), &schools)
		assertDeepEqual(t, "schools", sc.wantSchools, schools)

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

	var summary Summary
	getJSON(t, appServer.URL+"/summary", &summary)
	assertDeepEqual(t, "summary", scenarios[len(scenarios)-1].wantSummary, summary)
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

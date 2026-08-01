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

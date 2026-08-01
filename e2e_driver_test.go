//go:build e2e

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
)

// TestImportGoldenPath drives the real, unmodified NewServer handler
// through the whole workflow - import a tournament from the fake Tabroom
// server, then read it back via every read endpoint - and checks both the
// JSON API responses and what actually landed in the published GCS bucket.
// This is the test the whole e2e harness (fake Tabroom, testcontainers
// Postgres, fake-gcs-server) exists to support.
func TestImportGoldenPath(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	fixture, err := os.ReadFile("testdata/e2e/fixture_tournament.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	tabroomServer := newFakeTabroomServer(fakeTabroomTournament{
		id:       99001,
		date:     "2026-08-01",
		name:     "Fixture Debate Invitational",
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

	postImport(t, appServer.URL+"/tournaments/99001/import")

	pairings := TournamentPairings{}
	getJSON(t, appServer.URL+"/tournaments/99001/pairings/latest", &pairings)
	assertPairings(t, pairings)

	status := TournamentSchoolsStatus{}
	getJSON(t, appServer.URL+"/tournaments/99001/schools/status", &status)
	assertSchoolsStatus(t, status)

	summary := Summary{}
	getJSON(t, appServer.URL+"/summary", &summary)
	if summary.TournamentCount != 1 {
		t.Errorf("summary.TournamentCount = %d, want 1", summary.TournamentCount)
	}
	if summary.RoundCount != 1 {
		t.Errorf("summary.RoundCount = %d, want 1", summary.RoundCount)
	}

	assertPublishedPages(t, ctx)
}

func assertPairings(t *testing.T, pairings TournamentPairings) {
	t.Helper()

	if pairings.Name != "Fixture Debate Invitational" {
		t.Errorf("pairings.Name = %q", pairings.Name)
	}
	if pairings.UpdateTime != "2026-08-01 1:05PM" {
		t.Errorf("pairings.UpdateTime = %q, want %q", pairings.UpdateTime, "2026-08-01 1:05PM")
	}
	if len(pairings.EventPairings) != 1 {
		t.Fatalf("got %d event pairings, want 1: %+v", len(pairings.EventPairings), pairings.EventPairings)
	}

	event := pairings.EventPairings[0]
	if event.Name != "Public Forum" {
		t.Errorf("event.Name = %q, want %q", event.Name, "Public Forum")
	}
	if event.Number != 1 {
		t.Errorf("event.Number = %d, want 1", event.Number)
	}
	if event.Flighted {
		t.Errorf("event.Flighted = true, want false (single flight)")
	}
	if event.StartTime != "4:00AM" {
		t.Errorf("event.StartTime = %q, want %q", event.StartTime, "4:00AM")
	}
	if len(event.Pairings) != 1 {
		t.Fatalf("got %d pairings, want 1: %+v", len(event.Pairings), event.Pairings)
	}

	p := event.Pairings[0]
	if p.Room == nil || *p.Room != "Room 101" {
		t.Errorf("p.Room = %v, want %q", p.Room, "Room 101")
	}
	if p.AffEntry == nil || p.AffEntry.Name != "AA1" {
		t.Errorf("p.AffEntry = %+v, want code AA1", p.AffEntry)
	}
	if p.NegEntry == nil || p.NegEntry.Name != "BB1" {
		t.Errorf("p.NegEntry = %+v, want code BB1", p.NegEntry)
	}
	if p.AffResult == nil || *p.AffResult != WIN {
		t.Errorf("p.AffResult = %v, want WIN", p.AffResult)
	}
	if p.NegResult == nil || *p.NegResult != LOSS {
		t.Errorf("p.NegResult = %v, want LOSS", p.NegResult)
	}
	if len(p.Judges) != 1 || p.Judges[0].Name != "Jane Judge" || !p.Judges[0].Started {
		t.Errorf("p.Judges = %+v, want one started judge named Jane Judge", p.Judges)
	}
}

func assertSchoolsStatus(t *testing.T, status TournamentSchoolsStatus) {
	t.Helper()

	if len(status.SchoolsStatus) != 2 {
		t.Fatalf("got %d schools, want 2: %+v", len(status.SchoolsStatus), status.SchoolsStatus)
	}
	// GetSchoolStatus orders by school_name, so Alpha sorts before Beta.
	if status.SchoolsStatus[0].Name != "Alpha High" || !status.SchoolsStatus[0].CheckedIn {
		t.Errorf("SchoolsStatus[0] = %+v, want Alpha High checked in", status.SchoolsStatus[0])
	}
	if status.SchoolsStatus[1].Name != "Beta High" || status.SchoolsStatus[1].CheckedIn {
		t.Errorf("SchoolsStatus[1] = %+v, want Beta High not checked in", status.SchoolsStatus[1])
	}
}

func assertPublishedPages(t *testing.T, ctx context.Context) {
	t.Helper()

	storageClient, err := getStorageClient(ctx)
	if err != nil {
		t.Fatalf("getStorageClient: %v", err)
	}
	defer storageClient.Close()

	pairingsHTML := readObject(t, ctx, storageClient, "pairings.html")
	for _, want := range []string{"AA1", "BB1", "Room 101", "Jane Judge", "Public Forum Round #1"} {
		if !strings.Contains(pairingsHTML, want) {
			t.Errorf("pairings.html missing %q", want)
		}
	}

	statusHTML := readObject(t, ctx, storageClient, "status.html")
	for _, want := range []string{"Alpha High", "Beta High"} {
		if !strings.Contains(statusHTML, want) {
			t.Errorf("status.html missing %q", want)
		}
	}
}

func postImport(t *testing.T, url string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Fatalf("build import request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("import request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("import status = %d, want %d: %s", resp.StatusCode, http.StatusCreated, body)
	}
}

func getJSON(t *testing.T, url string, v any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s: status = %d: %s", url, resp.StatusCode, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("GET %s: decode: %v", url, err)
	}
}

func readObject(t *testing.T, ctx context.Context, client *storage.Client, name string) string {
	t.Helper()
	r, err := client.Bucket(gcsBucketName).Object(name).NewReader(ctx)
	if err != nil {
		t.Fatalf("read object %s: %v", name, err)
	}
	defer r.Close()
	content, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read object %s: %v", name, err)
	}
	return string(content)
}

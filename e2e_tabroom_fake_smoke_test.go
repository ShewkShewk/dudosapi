//go:build e2e

package main

import (
	"os"
	"testing"
	"time"

	"github.com/ShewkShewk/tbapi"
)

// TestFakeTabroomServer proves the fake server + fixture are actually
// compatible with the real tbapi client, before anything in dudosapi's own
// import path is layered on top. If this fails, the fixture or the fake
// server's routes/HTML/regex-matchable shape is wrong, not dudosapi.
func TestFakeTabroomServer(t *testing.T) {
	fixture, err := os.ReadFile("testdata/e2e/fixture_tournament.json")
	if err != nil {
		t.Fatalf("unable to read fixture: %v", err)
	}

	server := newFakeTabroomServer(fakeTabroomTournament{
		id:       99001,
		date:     "2026-08-01",
		name:     "Fixture Debate Invitational",
		dataJSON: fixture,
	})
	defer server.Close()

	api, err := tbapi.NewBuilder().
		WithHostname(server.URL).
		WithUsername("test-user").
		WithPassword("test-password").
		Build()
	if err != nil {
		t.Fatalf("unable to build tbapi client: %v", err)
	}

	tournaments, err := api.GetTournaments()
	if err != nil {
		t.Fatalf("GetTournaments: %v", err)
	}
	if len(tournaments) != 1 {
		t.Fatalf("GetTournaments: got %d tournaments, want 1: %+v", len(tournaments), tournaments)
	}
	got := tournaments[0]
	if got.Id != 99001 {
		t.Errorf("GetTournaments: Id = %d, want 99001", got.Id)
	}
	if got.Name != "Fixture Debate Invitational" {
		t.Errorf("GetTournaments: Name = %q, want %q", got.Name, "Fixture Debate Invitational")
	}
	wantDate, _ := time.Parse(time.DateOnly, "2026-08-01")
	if !got.Date.Equal(wantDate) {
		t.Errorf("GetTournaments: Date = %v, want %v", got.Date, wantDate)
	}

	data, err := api.GetTournamentData(99001)
	if err != nil {
		t.Fatalf("GetTournamentData: %v", err)
	}
	if data.Name != "Fixture Debate Invitational" {
		t.Errorf("GetTournamentData: Name = %q, want %q", data.Name, "Fixture Debate Invitational")
	}
	if len(data.Schools) != 2 {
		t.Fatalf("GetTournamentData: got %d schools, want 2", len(data.Schools))
	}
	if data.Schools[0].Name != "Alpha High" {
		t.Errorf("GetTournamentData: Schools[0].Name = %q, want %q", data.Schools[0].Name, "Alpha High")
	}
}

//go:build e2e

package main

// e2eScenario describes one full import-and-verify pass: which fixture the
// fake Tabroom server should serve, and what the app is expected to produce
// for it. New scenarios (byes, forfeits, re-import, ...) are added to
// TestImportScenarios without touching runScenario or the comparison
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

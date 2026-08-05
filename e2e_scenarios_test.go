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
		wantStatusHTMLContains: []string{
			`<td style="color: green; font-weight: bold;">Alpha High</td>`,
			`<td style="color: red; font-weight: bold;">Beta High</td>`,
		},
	}
}

// multiEventMultiRoundScenario: one tournament, two schools, two event types
// (Public Forum, Lincoln Douglas) each with four rounds. All four rounds per
// event are published, but /pairings/latest only surfaces the highest
// published round number per event - so the expected pairings reflect round
// #4 even though rounds #1-#3 exist in the fixture. Summary.RoundCount,
// unlike the pairings endpoint, counts every started section across all
// rounds and both events (4 rounds * 2 events = 8), which is why it doesn't
// match either event's round Number.
func multiEventMultiRoundScenario() e2eScenario {
	return e2eScenario{
		name:           "multiple event types with four rounds each",
		tournamentID:   99002,
		tournamentDate: "2026-08-01",
		tournamentName: "Multi-Event Multi-Round Invitational",
		fixturePath:    "testdata/e2e/fixture_multi_event_multi_round.json",

		wantPairings: TournamentPairings{
			Name:       "Multi-Event Multi-Round Invitational",
			UpdateTime: "2026-08-01 1:05PM",
			EventPairings: []EventPairing{
				{
					Name:      "Lincoln Douglas",
					Number:    4,
					Flighted:  false,
					StartTime: "5:45AM",
					Pairings: []Pairing{
						{
							SectionId: 8,
							Flight:    1,
							Room:      ptr("Room 102"),
							AffEntry:  &Entry{Id: 11, Name: "AA2"},
							AffResult: ptr(LOSS),
							NegEntry:  &Entry{Id: 21, Name: "BB2"},
							NegResult: ptr(WIN),
							Judges: []Judge{
								{Id: 1, PersonId: 501, Name: "Jane Judge", Started: true},
							},
						},
					},
				},
				{
					Name:      "Public Forum",
					Number:    4,
					Flighted:  false,
					StartTime: "5:30AM",
					Pairings: []Pairing{
						{
							SectionId: 4,
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
			Name:       "Multi-Event Multi-Round Invitational",
			UpdateTime: "2026-08-01 1:05PM",
			SchoolsStatus: []SchoolStatus{
				{Id: 1, Name: "Alpha High", CheckedIn: true},
				{Id: 2, Name: "Beta High", CheckedIn: false},
			},
		},

		wantSummary: Summary{TournamentCount: 1, RoundCount: 8},

		wantPairingsHTMLContains: []string{
			"AA1", "BB1", "AA2", "BB2", "Room 101", "Room 102", "Jane Judge",
			"Public Forum Round #4", "Lincoln Douglas Round #4",
		},
		wantStatusHTMLContains: []string{
			`<td style="color: green; font-weight: bold;">Alpha High</td>`,
			`<td style="color: red; font-weight: bold;">Beta High</td>`,
		},
	}
}

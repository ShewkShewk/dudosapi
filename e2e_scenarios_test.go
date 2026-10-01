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

	wantPairings          TournamentPairings
	wantSchoolsStatus     TournamentSchoolsStatus
	wantEventSchoolCounts TournamentEventSchoolCounts
	wantSummary           Summary
	wantSchools           []School

	wantPairingsHTMLContains []string
	wantStatusHTMLContains   []string
}

// e2eScenarioSequence names an ordered group of e2eScenarios that get
// imported, in order, against one shared database within a single subtest -
// see runScenario. A sequence of one is an ordinary single-tournament
// scenario; a longer sequence exercises cross-tournament behavior, such as
// a later import overwriting the previously published pairings.html/
// status.html, or Summary's counts being global rather than scoped to one
// tournament.
type e2eScenarioSequence struct {
	name      string
	scenarios []e2eScenario
}

// solo wraps a single scenario as a one-element sequence, reusing the
// scenario's own name as the subtest name.
func solo(sc e2eScenario) e2eScenarioSequence {
	return e2eScenarioSequence{name: sc.name, scenarios: []e2eScenario{sc}}
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

		wantEventSchoolCounts: TournamentEventSchoolCounts{
			Name: "Fixture Debate Invitational",
			Events: []EventSchoolCounts{
				{
					Id: 1, Name: "Public Forum", EntryCount: 2, StudentCount: 2,
					Schools: []SchoolEntryCount{
						{Id: 1, Name: "Alpha High", EntryCount: 1, StudentCount: 1},
						{Id: 2, Name: "Beta High", EntryCount: 1, StudentCount: 1},
					},
				},
			},
		},

		wantSummary: Summary{TournamentCount: 1, RoundCount: 1},

		wantSchools: []School{
			{
				Id:   1,
				Name: "Alpha High",
			},
			{
				Id:   2,
				Name: "Beta High",
			},
		},

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
							SectionId: 108,
							Flight:    1,
							Room:      ptr("Room 102"),
							AffEntry:  &Entry{Id: 111, Name: "AA2"},
							AffResult: ptr(LOSS),
							NegEntry:  &Entry{Id: 121, Name: "BB2"},
							NegResult: ptr(WIN),
							Judges: []Judge{
								{Id: 101, PersonId: 501, Name: "Jane Judge", Started: true},
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
							SectionId: 104,
							Flight:    1,
							Room:      ptr("Room 101"),
							AffEntry:  &Entry{Id: 110, Name: "AA1"},
							AffResult: ptr(WIN),
							NegEntry:  &Entry{Id: 120, Name: "BB1"},
							NegResult: ptr(LOSS),
							Judges: []Judge{
								{Id: 101, PersonId: 501, Name: "Jane Judge", Started: true},
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

		wantEventSchoolCounts: TournamentEventSchoolCounts{
			Name: "Multi-Event Multi-Round Invitational",
			Events: []EventSchoolCounts{
				{
					Id: 102, Name: "Lincoln Douglas", EntryCount: 2, StudentCount: 2,
					Schools: []SchoolEntryCount{
						{Id: 1, Name: "Alpha High", EntryCount: 1, StudentCount: 1},
						{Id: 2, Name: "Beta High", EntryCount: 1, StudentCount: 1},
					},
				},
				{
					Id: 101, Name: "Public Forum", EntryCount: 2, StudentCount: 2,
					Schools: []SchoolEntryCount{
						{Id: 1, Name: "Alpha High", EntryCount: 1, StudentCount: 1},
						{Id: 2, Name: "Beta High", EntryCount: 1, StudentCount: 1},
					},
				},
			},
		},

		wantSummary: Summary{TournamentCount: 1, RoundCount: 8},

		wantSchools: []School{
			{
				Id:   1,
				Name: "Alpha High",
			},
			{
				Id:   2,
				Name: "Beta High",
			},
		},

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

// multipleTournamentsScenario imports the golden path tournament followed by
// the multi-event tournament against one shared database. Each scenario's
// pairings/schools-status/HTML are still asserted immediately after its own
// import, which also proves the second import overwrites the first's
// published pairings.html/status.html. wantSummary on the second scenario is
// overridden to the totals across both tournaments, since RoundCount and
// TournamentCount aggregate globally rather than being scoped to one
// tournament - runScenario only checks the last scenario's wantSummary when
// a sequence has more than one entry.
func multipleTournamentsScenario() e2eScenarioSequence {
	second := multiEventMultiRoundScenario()
	second.wantSummary = Summary{TournamentCount: 2, RoundCount: 9}
	return e2eScenarioSequence{
		name:      "multiple tournaments imported in sequence",
		scenarios: []e2eScenario{goldenPathScenario(), second},
	}
}

// entryCountsScenario: one tournament with no rounds, built to pin down how
// /events/schools counts entries. Policy has a two-student entry (310), a
// hybrid entry whose students come from both schools (311), and an inactive
// entry (320); World Schools has one two-student entry (321); Original
// Oratory is a speech event, which the import skips entirely (330).
//
//   - 320 is inactive, so Beta High's Policy row shows only the hybrid entry.
//   - 311 counts once for each school, so the per-school Policy entry counts
//     (2 + 1) sum to more than the event's distinct-entry total (2).
//   - 330's event isn't imported, so it appears nowhere.
func entryCountsScenario() e2eScenario {
	return e2eScenario{
		name:           "entry counts by event and school",
		tournamentID:   99003,
		tournamentDate: "2026-08-01",
		tournamentName: "Entry Count Invitational",
		fixturePath:    "testdata/e2e/fixture_entry_counts.json",

		wantPairings: TournamentPairings{
			Name:          "Entry Count Invitational",
			UpdateTime:    "2026-08-01 1:05PM",
			EventPairings: []EventPairing{},
		},

		wantSchoolsStatus: TournamentSchoolsStatus{
			Name:       "Entry Count Invitational",
			UpdateTime: "2026-08-01 1:05PM",
			SchoolsStatus: []SchoolStatus{
				{Id: 1, Name: "Alpha High", CheckedIn: true},
				{Id: 2, Name: "Beta High", CheckedIn: false},
			},
		},

		wantEventSchoolCounts: TournamentEventSchoolCounts{
			Name: "Entry Count Invitational",
			Events: []EventSchoolCounts{
				{
					Id: 301, Name: "Policy", EntryCount: 2, StudentCount: 4,
					Schools: []SchoolEntryCount{
						{Id: 1, Name: "Alpha High", EntryCount: 2, StudentCount: 3},
						{Id: 2, Name: "Beta High", EntryCount: 1, StudentCount: 1},
					},
				},
				{
					Id: 302, Name: "World Schools", EntryCount: 1, StudentCount: 2,
					Schools: []SchoolEntryCount{
						{Id: 2, Name: "Beta High", EntryCount: 1, StudentCount: 2},
					},
				},
			},
		},

		wantSummary: Summary{TournamentCount: 1, RoundCount: 0},

		wantSchools: []School{
			{
				Id:   1,
				Name: "Alpha High",
			},
			{
				Id:   2,
				Name: "Beta High",
			},
		},

		wantStatusHTMLContains: []string{
			`<td style="color: green; font-weight: bold;">Alpha High</td>`,
			`<td style="color: red; font-weight: bold;">Beta High</td>`,
		},
	}
}

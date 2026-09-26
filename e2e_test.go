//go:build e2e

package main

import "testing"

func TestImportScenarios(t *testing.T) {
	sequences := []e2eScenarioSequence{
		solo(goldenPathScenario()),
		solo(multiEventMultiRoundScenario()),
		multipleTournamentsScenario(),
		solo(entryCountsScenario()),
	}

	for _, seq := range sequences {
		t.Run(seq.name, func(t *testing.T) {
			runScenario(t, seq.scenarios...)
		})
	}
}

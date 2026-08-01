//go:build e2e

package main

import "testing"

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

package main

import (
	"context"
	"log"

	"github.com/ShewkShewk/dudosapi/internal/db/sqlc"
)

func getSummary(ctx context.Context, queries *sqlc.Queries) (*Summary, error) {
	tournamentCount, err := queries.GetTournamentCount(ctx)
	if err != nil {
		log.Printf("unable to get tournament count %v", err)
		return nil, err
	}

	return &Summary{
		TournamentCount: tournamentCount,
	}, nil
}

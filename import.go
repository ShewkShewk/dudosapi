package main

import (
	"context"
	"log"
	"strconv"

	"github.com/ShewkShewk/dudosapi/internal/db/sqlc"
	"github.com/ShewkShewk/tbapi"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func importTournament(ctx context.Context, conn *pgxpool.Pool, queries *sqlc.Queries, tourn *tbapi.TournamentData) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Printf("importTournament: unable to open transaction for tournament %v %v", tourn.Name, err)
		return err
	}
	defer tx.Rollback(ctx)
	qtx := queries.WithTx(tx)
	err = importSchools(ctx, qtx, tourn)
	if err != nil {
		log.Printf("importTournament: unable to import schools for %v %v", tourn.Name, err)
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("importTournament: unable to commit import for %v %v", tourn.Name, err)
		return err
	}
	return nil
}

func importSchools(ctx context.Context, qtx *sqlc.Queries, tourn *tbapi.TournamentData) error {
	for _, school := range tourn.Schools {
		schoolId, err := strconv.Atoi(school.Id)
		if err != nil {
			log.Printf("importSchools: unable to convert %s to school id.", school.Id)
			continue
		}
		err = qtx.InsertSchool(ctx, sqlc.InsertSchoolParams{
			ID: int32(schoolId),
			Name: pgtype.Text{
				String: school.Name,
				Valid:  true,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

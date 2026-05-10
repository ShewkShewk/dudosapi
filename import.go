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

func importTournament(ctx context.Context, conn *pgxpool.Pool, queries *sqlc.Queries, tournId int32, tourn *tbapi.TournamentData) error {
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
	err = importStudents(ctx, qtx, tourn)
	if err != nil {
		log.Printf("importTournament: unable to import students for %v %v", tourn.Name, err)
		return err
	}
	err = importEvents(ctx, qtx, tournId, tourn)
	if err != nil {
		log.Printf("importTournament: unable to import events for %v %v", tourn.Name, err)
		return err
	}
	err = importEntries(ctx, qtx, tournId, tourn)
	if err != nil {
		log.Printf("importTournament: unable to import entries for %v %v", tourn.Name, err)
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
		err := qtx.InsertSchool(ctx, sqlc.InsertSchoolParams{
			ID: int32(school.Chapter),
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

func importEvents(ctx context.Context, qtx *sqlc.Queries, tournId int32, tourn *tbapi.TournamentData) error {
	for _, category := range tourn.Categories {
		for _, event := range category.Events {
			eventId, err := strconv.Atoi(event.Id)
			if err != nil {
				log.Printf("importEvents: unable to convert %s to event id", event.Id)
				return err
			}
			err = qtx.InsertEvent(ctx, sqlc.InsertEventParams{
				ID: int32(eventId),
				TournamentID: pgtype.Int4{
					Int32: tournId,
					Valid: true,
				},
				Name: pgtype.Text{
					String: event.Name,
					Valid:  true,
				},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func importStudents(ctx context.Context, qtx *sqlc.Queries, tourn *tbapi.TournamentData) error {
	for _, school := range tourn.Schools {
		for _, student := range school.Students {
			studentId, err := strconv.Atoi(student.Id)
			if err != nil {
				log.Printf("importStudents: unable to convert studentId %v to int. %v", student.Id, err)
				return err
			}
			err = qtx.InsertStudent(ctx, sqlc.InsertStudentParams{
				ID: int32(studentId),
				SchoolID: pgtype.Int4{
					Int32: int32(school.Chapter),
					Valid: true,
				},
				FirstName:  student.First,
				MiddleName: student.Middle,
				LastName:   student.Last,
				GradYear:   int32(student.GradYear),
			})
			if err != nil {
				log.Printf("importStudent unable to insert student %v %v", student.Id, err)
				return err
			}
		}
	}
	return nil
}

func importEntries(ctx context.Context, qtx *sqlc.Queries, tournID int32, tourn *tbapi.TournamentData) error {
	for _, school := range tourn.Schools {
		for _, entry := range school.Entries {
			entryId, err := strconv.Atoi(entry.Id)
			if err != nil {
				log.Printf("importEntries: unable to convert entryId to int %v %v", entry.Id, err)
				return err
			}
			err = qtx.InsertEntry(ctx, sqlc.InsertEntryParams{
				ID: pgtype.Int4{
					Int32: int32(entryId),
					Valid: true,
				},
				TournamentID: pgtype.Int4{
					Int32: tournID,
					Valid: true,
				},
				EventID: pgtype.Int4{
					Int32: int32(entry.Event),
					Valid: true,
				},
				Code: pgtype.Text{
					String: entry.Code,
					Valid:  true,
				},
				Active: pgtype.Bool{
					Bool:  entry.Active == 1,
					Valid: true,
				},
			})
			if err != nil {
				log.Printf("importEntries: unable to insert entry for %v %v", entryId, err)
				return err
			}
		}
	}
	return nil
}

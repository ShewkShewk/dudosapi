package main

import (
	"context"
	"log"
	"strconv"
	"time"

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
	err = importSitesAndRooms(ctx, qtx, tourn)
	if err != nil {
		log.Printf("importTOurnament: unable to import rounds for %v %v", tourn.Name, err)
		return err
	}
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
	err = importRounds(ctx, qtx, tourn)
	if err != nil {
		log.Printf("importTournament: unable to import rounds for %v %v", tourn.Name, err)
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("importTournament: unable to commit import for %v %v", tourn.Name, err)
		return err
	}
	return nil
}

func importSitesAndRooms(ctx context.Context, qtx *sqlc.Queries, tourn *tbapi.TournamentData) error {
	siteBatch := make([]sqlc.InsertSitesParams, len(tourn.Sites))
	var roomBatch []sqlc.InsertRoomsParams
	for i, site := range tourn.Sites {
		siteId, err := strconv.Atoi(site.Id)
		if err != nil {
			log.Printf("importSitesAndRooms: unable to convert site id %v to int", site.Id)
			return err
		}
		siteBatch[i] = sqlc.InsertSitesParams{
			ID: int32(siteId),
			Name: pgtype.Text{
				String: site.Name,
				Valid:  true,
			},
		}
		for _, room := range site.Rooms {
			roomId, err := strconv.Atoi(room.Id)
			if err != nil {
				log.Printf("importSitesAndRooms: unable to convert room id %v to int", room.Id)
				return err
			}
			roomBatch = append(roomBatch, sqlc.InsertRoomsParams{
				ID: int32(roomId),
				SiteID: pgtype.Int4{
					Int32: int32(siteId),
					Valid: true,
				},
				Name: pgtype.Text{
					String: room.Name,
					Valid:  true,
				},
			})
		}
	}
	siteResults := qtx.InsertSites(ctx, siteBatch)
	err := batchExecErr(siteResults.Exec, siteResults.Close)
	if err != nil {
		log.Printf("importSitesAndRooms: unable to insert sites for tourn %v", tourn.Name)
		return err
	}
	roomResults := qtx.InsertRooms(ctx, roomBatch)
	return batchExecErr(roomResults.Exec, roomResults.Close)
}

func importSchools(ctx context.Context, qtx *sqlc.Queries, tourn *tbapi.TournamentData) error {
	batch := make([]sqlc.InsertSchoolParams, len(tourn.Schools))
	for i, school := range tourn.Schools {
		batch[i] = sqlc.InsertSchoolParams{
			ID: int32(school.Chapter),
			Name: pgtype.Text{
				String: school.Name,
				Valid:  true,
			},
		}
	}
	results := qtx.InsertSchool(ctx, batch)
	return batchExecErr(results.Exec, results.Close)
}

func importEvents(ctx context.Context, qtx *sqlc.Queries, tournId int32, tourn *tbapi.TournamentData) error {
	var batch []sqlc.InsertEventParams
	for _, category := range tourn.Categories {
		for _, event := range category.Events {
			eventId, err := strconv.Atoi(event.Id)
			if err != nil {
				log.Printf("importEvents: unable to convert %s to event id", event.Id)
				return err
			}
			batch = append(batch, sqlc.InsertEventParams{
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
		}
	}
	results := qtx.InsertEvent(ctx, batch)
	return batchExecErr(results.Exec, results.Close)
}

func importStudents(ctx context.Context, qtx *sqlc.Queries, tourn *tbapi.TournamentData) error {
	var batch []sqlc.InsertStudentParams
	for _, school := range tourn.Schools {
		for _, student := range school.Students {
			studentId, err := strconv.Atoi(student.Id)
			if err != nil {
				log.Printf("importStudents: unable to convert studentId %v to int. %v", student.Id, err)
				return err
			}
			batch = append(batch, sqlc.InsertStudentParams{
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
		}
	}
	results := qtx.InsertStudent(ctx, batch)
	return batchExecErr(results.Exec, results.Close)
}

func importEntries(ctx context.Context, qtx *sqlc.Queries, tournID int32, tourn *tbapi.TournamentData) error {
	var entryBatch []sqlc.InsertEntryParams
	var studentEntryBatch []sqlc.InsertStudentEntriesParams
	for _, school := range tourn.Schools {
		for _, entry := range school.Entries {
			entryId, err := strconv.Atoi(entry.Id)
			if err != nil {
				log.Printf("importEntries: unable to convert entryId to int %v %v", entry.Id, err)
				return err
			}
			entryBatch = append(entryBatch, sqlc.InsertEntryParams{
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
			for _, student := range entry.Students {
				studentId, err := strconv.Atoi(student)
				if err != nil {
					log.Printf("importEntries: unable to convert student id %v within entry %v", studentId, entryId)
					return err
				}
				studentEntryBatch = append(studentEntryBatch, sqlc.InsertStudentEntriesParams{
					StudentID: int32(studentId),
					EntryID:   int32(entryId),
				})
			}
		}
	}
	entryResults := qtx.InsertEntry(ctx, entryBatch)
	err := batchExecErr(entryResults.Exec, entryResults.Close)
	if err != nil {
		log.Printf("importEntries: unable to import entries for tournId: %v", tournID)
		return err
	}
	studentEntryResults := qtx.InsertStudentEntries(ctx, studentEntryBatch)
	return batchExecErr(studentEntryResults.Exec, studentEntryResults.Close)
}

func importRounds(ctx context.Context, qtx *sqlc.Queries, tourn *tbapi.TournamentData) error {
	var batch []sqlc.InsertRoundParams
	for _, category := range tourn.Categories {
		for _, event := range category.Events {
			eventId, err := strconv.Atoi(event.Id)
			if err != nil {
				log.Printf("importRounds: unable to convert event id %v to int", eventId)
				return err
			}
			for _, round := range event.Rounds {
				roundId, err := strconv.Atoi(round.Id)
				if err != nil {
					log.Printf("importRounds: unable to convert round id %v to int", roundId)
					return err
				}
				parsedTime, err := time.Parse("2006-01-02 15:04:05", round.StartTime)
				if err != nil {
					log.Printf("importRounds: unable to convert timestamp %v to Timestamp.", round.StartTime)
					return err
				}
				batch = append(batch, sqlc.InsertRoundParams{
					ID: int32(roundId),
					EventID: pgtype.Int4{
						Int32: int32(eventId),
						Valid: true,
					},
					Number: int32(round.Name),
					StartTime: pgtype.Timestamp{
						Time:  parsedTime,
						Valid: true,
					},
					Published: round.Published == 1,
				})
			}
		}
	}
	results := qtx.InsertRound(ctx, batch)
	return batchExecErr(results.Exec, results.Close)
}

func batchExecErr(exec func(func(int, error)), close func() error) error {
	var batchErr error

	exec(func(i int, err error) {
		if err != nil && batchErr == nil {
			batchErr = err
		}
	})

	closeErr := close()

	if batchErr != nil {
		return batchErr
	}

	return closeErr
}

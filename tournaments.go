package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ShewkShewk/dudosapi/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getLatestPairings(ctx context.Context, conn *pgxpool.Pool, queries *sqlc.Queries, tournId int32) (*TournamentPairings, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Printf("getLatestPairings: unable to open trasnaction for tournament %v %v", tournId, err)
		return nil, err
	}
	defer tx.Rollback(ctx)
	qtx := queries.WithTx(tx)
	tournamentData, err := qtx.GetTournament(ctx, tournId)
	if err != nil {
		log.Printf("getLatestPairings: unable to get tournament data for %v %v", tournId, err)
		return nil, err
	}
	if !tournamentData.Name.Valid || !tournamentData.Date.Valid || !tournamentData.UpdatedTime.Valid {
		log.Printf("getLatestPairings: invalid data returned for %v", tournId)
		return nil, err
	}
	tournamentName := tournamentData.Name.String
	updateTime, err := utcToCentralTime(tournamentData.UpdatedTime.String)
	if err != nil {
		log.Printf("getLatestPairings: unable to convert time %v to timezone for tournament %v", tournamentData.UpdatedTime.String, tournId)
		return nil, err
	}
	rows, err := qtx.GetLatestPublishedRoundsPerEvent(ctx, pgtype.Int4{
		Int32: tournId,
		Valid: true,
	})
	if err != nil {
		log.Printf("getLatestPairings: unable to get latest published rounds for %v %v", tournId, err)
		return nil, err
	}
	roundIds := make([]int32, len(rows))
	for i, row := range rows {
		roundIds[i] = row.RoundID
	}
	dbPairings, err := qtx.GetPairingsWithBallots(ctx, roundIds)
	if err != nil {
		log.Printf("getLatestPairings: unable to get pairings with ballots for tournament %v with sections %v %v", tournId, roundIds, err)
		return nil, err
	}
	pairingByEntry := make(map[int32][]Pairing)
	flightedByEntry := make(map[int32]struct{})
	for _, dbPairing := range dbPairings {
		pairing, err := toPairing(dbPairing)
		if err != nil {
			log.Printf("getLatestPairing: unable to convert db pairing for section %v to domain object %v", dbPairing.SectionID, err)
			return nil, err
		}
		eventId := dbPairing.EventID
		pairingByEntry[eventId] = append(pairingByEntry[eventId], *pairing)
		if pairing.Flight > 1 {
			flightedByEntry[eventId] = struct{}{}
		}
	}
	eventPairings := make([]EventPairing, len(rows))
	for i, row := range rows {
		eventId := row.EventID.Int32
		_, flighted := flightedByEntry[eventId]
		eventPairings[i] = EventPairing{
			Name:      row.EventName.String,
			Number:    int(row.RoundNumber),
			Flighted:  flighted,
			StartTime: row.StartTime.Time.In(getTimezone()).Format(time.Kitchen),
			Pairings:  pairingByEntry[eventId],
		}
	}

	return &TournamentPairings{
		Name:          tournamentName,
		UpdateTime:    updateTime,
		EventPairings: eventPairings,
	}, nil
}

func toPairing(row sqlc.GetPairingsWithBallotsRow) (*Pairing, error) {
	var sectionId int
	if row.SectionID.Valid {
		sectionId = int(row.SectionID.Int32)
	}
	flight := int(row.Flight)
	var room *string
	if row.RoomName.Valid {
		room = &row.RoomName.String
	}
	var affEntry *Entry
	if row.AffTeamEntryCode.Valid && row.AffTeamEntryID.Valid {
		affEntry = &Entry{
			Id:   int(row.AffTeamEntryID.Int32),
			Name: row.AffTeamEntryCode.String,
		}
	}
	var negEntry *Entry
	if row.NegTeamEntryCode.Valid && row.NegTeamEntryID.Valid {
		negEntry = &Entry{
			Id:   int(row.NegTeamEntryID.Int32),
			Name: row.NegTeamEntryCode.String,
		}
	}
	var judges []Judge
	if row.AssociatedJudges != nil {
		var associatedJudges []sqlc.AssociatedJudge
		err := json.Unmarshal(row.AssociatedJudges, &associatedJudges)
		if err != nil {
			log.Printf("toPairing: unable to convert %v to judge array for section %v %v", row.AssociatedJudges, row.SectionID, err)
			return nil, err
		}
		for _, associatedJudge := range associatedJudges {
			judges = append(judges, Judge{
				Id:       associatedJudge.Id,
				PersonId: associatedJudge.PersonId,
				Name:     fmt.Sprintf("%s %s", associatedJudge.FirstName, associatedJudge.LastName),
				Started:  associatedJudge.Started,
			})
		}
	}

	affResult := UNKNOWN
	negResult := UNKNOWN
	affWinBallotCount := 0
	negWinBallotCount := 0
	if row.AffTeamBallots != nil {
		ballotResults, err := getBallotResults(row.AffTeamBallots)
		if err != nil {
			log.Printf("toPairing: unable to convert %v to ballotResults from aff %v", row.AffTeamBallots, err)
			return nil, err
		}
		for _, result := range ballotResults {
			switch result {
			case WIN:
				affWinBallotCount += 1
			case BYE, FFT:
				affResult = result
			}
		}
	}

	if row.NegTeamBallots != nil {
		ballotResults, err := getBallotResults(row.NegTeamBallots)
		if err != nil {
			log.Printf("toPairing: unable to convert %v to ballotResults from neg %v", row.NegTeamBallots, err)
			return nil, err
		}
		for _, result := range ballotResults {
			switch result {
			case WIN:
				negWinBallotCount += 1
			case BYE, FFT:
				negResult = result
			}
		}
	}
	if affWinBallotCount != negWinBallotCount {
		if affWinBallotCount > negWinBallotCount {
			affResult = WIN
			negResult = LOSS
		} else {
			affResult = LOSS
			negResult = WIN
		}
	}
	var affResultPtr *BallotResult
	var negResultPtr *BallotResult
	if affResult != UNKNOWN {
		affResultPtr = &affResult
	}
	if negResult != UNKNOWN {
		negResultPtr = &negResult
	}
	pairing := Pairing{
		SectionId: sectionId,
		Flight:    flight,
		Room:      room,
		AffEntry:  affEntry,
		AffResult: affResultPtr,
		NegEntry:  negEntry,
		NegResult: negResultPtr,
		Judges:    judges,
	}
	return &pairing, nil
}

func getBallotResults(content []byte) ([]BallotResult, error) {
	var ballots []sqlc.TeamBallot
	err := json.Unmarshal(content, &ballots)
	if err != nil {
		log.Printf("getBallotResult: unable to unmarshall %v to TeamBallot %v", content, err)
		return nil, nil
	}
	toReturn := make([]BallotResult, len(ballots))
	for i, ballot := range ballots {
		toReturn[i] = BallotResult(ballot.Result)
	}
	return toReturn, nil
}

func utcToCentralTime(utcTimeStr string) (string, error) {
	layout := "2006-01-02T15:04:05"
	utcTime, err := time.Parse(layout, utcTimeStr)
	if err != nil {
		return "", fmt.Errorf("error parsing time: %w", err)
	}
	centralTime := utcTime.In(getTimezone())
	return fmt.Sprintf("%v %v", centralTime.Format(time.DateOnly), centralTime.Format(time.Kitchen)), nil
}

func getTimezone() *time.Location {
	timeZone, _ := time.LoadLocation("America/Chicago")
	return timeZone
}

func getSideNamesFor(event string) (string, string) {
	eventToAffNegName := map[string][]string{
		"world":     {"Prop", "Opp"},
		"ws":        {"Prop", "Opp"},
		"pol":       {"Aff", "Neg"},
		"lincoln":   {"Aff", "Neg"},
		"community": {"Aff", "Neg"},
	}
	lowercasedEventName := strings.ToLower(event)
	for key, value := range eventToAffNegName {
		if strings.Contains(lowercasedEventName, key) {
			return value[0], value[1]
		}
	}
	return "Aff", "Neg"
}

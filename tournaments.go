package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
	for _, dbPairing := range dbPairings {
		pairing, err := toPairing(dbPairing)
		if err != nil {
			log.Printf("getLatestPairing: unable to convert db pairing for section %v to domain object %v", dbPairing.SectionID, err)
			return nil, err
		}
		eventId := dbPairing.EventID
		pairingByEntry[eventId] = append(pairingByEntry[eventId], *pairing)
	}
	eventPairings := make([]EventPairing, len(rows))
	for i, row := range rows {
		eventPairings[i] = EventPairing{
			Name:     row.EventName.String,
			Number:   int(row.RoundNumber),
			Pairings: pairingByEntry[row.EventID.Int32],
		}
	}

	return &TournamentPairings{
		Name:          "tournamentNameHere",
		UpdateTime:    "sometimehere",
		EventPairings: eventPairings,
	}, nil
}

func toPairing(row sqlc.GetPairingsWithBallotsRow) (*Pairing, error) {
	var room string
	if row.RoomName.Valid {
		room = row.RoomName.String
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

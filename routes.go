package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ShewkShewk/dudosapi/internal/db/sqlc"
	"github.com/ShewkShewk/tbapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(config *Config) (http.Handler, error) {
	mux := http.NewServeMux()
	ctx := context.Background()
	tb, err := getTabroomApi(config)
	if err != nil {
		return nil, err
	}
	queries, err := getDbExecutor(ctx, config)
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /tournaments", handleGetTournaments(tb, queries))
	mux.Handle("POST /tournaments/{id}/import", handleImportTournament(tb, queries))
	mux.Handle("DELETE /tournaments/{id}", handleDeleteTournament(queries))
	return mux, nil
}

func getTabroomApi(config *Config) (*tbapi.TabroomApi, error) {
	tabroomConfig := config.tabroomConfig
	built, err := tbapi.NewBuilder().
		WithHostname(tabroomConfig.hostname).
		WithUsername(tabroomConfig.username).
		WithPassword(tabroomConfig.password).
		Build()
	if err != nil {
		return nil, err
	}
	return built, nil
}

func getDbExecutor(ctx context.Context, config *Config) (*sqlc.Queries, error) {
	conn, err := pgxpool.New(ctx, config.dbConnectionString)
	if err != nil {
		return nil, err
	}
	return sqlc.New(conn), nil
}

func handleGetTournaments(tb *tbapi.TabroomApi, queries *sqlc.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbtourns, err := queries.GetLoadedTournaments(r.Context())
		if err != nil {
			log.Printf("handleGetTournaments SQL error %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		loadedTourns := make(map[int32]string)
		for _, t := range dbtourns {
			loadedTourns[t.ID] = t.Name.String
		}
		tbtourns, err := tb.GetTournaments()
		if err != nil {
			log.Printf("handleGetTournaments error from tbapi GetTournaments: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tournaments := make([]Tournament, len(tbtourns))
		for i, tourn := range tbtourns {
			_, loaded := loadedTourns[int32(tourn.Id)]
			tournaments[i] = Tournament{
				Id:     tourn.Id,
				Date:   tourn.Date.Format(time.DateOnly),
				Name:   tourn.Name,
				Loaded: loaded,
			}
		}
		err = encode(w, r, http.StatusOK, tournaments)
		if err != nil {
			log.Printf("handleGetTournaments error from encoding: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func handleImportTournament(tb *tbapi.TabroomApi, queries *sqlc.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}
		tournament, err := tb.GetTournamentData(id)
		if err != nil {
			log.Printf("handleImportTournaments: unable to download tournament %v", id)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		raw, err := json.Marshal(tournament)
		if err != nil {
			log.Printf("handleImportTournaments: unable to marshal tournament: %v", id)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = queries.LoadTournament(r.Context(), sqlc.LoadTournamentParams{
			ID:  int32(id),
			Raw: raw,
		})
		if err != nil {
			log.Printf("handleImportTournaments: unable to save tournament to db: %v", id)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func handleDeleteTournament(queries *sqlc.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}
		err = queries.DeleteTournament(r.Context(), int32(id))
		if err != nil {
			log.Printf("handleDeleteTournament: unable to delete %v", id)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

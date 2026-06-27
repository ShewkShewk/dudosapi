package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/ShewkShewk/dudosapi/internal/db/sqlc"
	"github.com/ShewkShewk/tbapi"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(config *Config) (http.Handler, error) {
	mux := http.NewServeMux()
	ctx := context.Background()
	tb, err := getTabroomApi(config)
	if err != nil {
		return nil, err
	}
	dbConn, err := getDbExecutor(ctx, config)
	if err != nil {
		log.Printf("unable to get db connection %v", err)
		return nil, err
	}
	storageClient, err := getStorageClient(ctx)
	if err != nil {
		log.Printf("unable to get storage client %v", err)
		return nil, err
	}
	queries := sqlc.New(dbConn)
	mux.Handle("GET /tournaments", handleGetTournaments(tb, queries))
	mux.Handle("POST /tournaments/{id}/import", handleImportTournament(tb, dbConn, queries, storageClient))
	mux.Handle("DELETE /tournaments/{id}", handleDeleteTournament(queries))
	mux.Handle("GET /tournaments/{id}/pairings/latest", handleGetLatestPairings(dbConn, queries))
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

func getDbExecutor(ctx context.Context, config *Config) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, config.dbConnectionString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getStorageClient(ctx context.Context) (*storage.Client, error) {
	client, err := storage.NewClient(ctx)
	return client, err
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
			loadedTourns[t.ID] = t.UpdatedTime.String
		}
		tbtourns, err := tb.GetTournaments()
		if err != nil {
			log.Printf("handleGetTournaments error from tbapi GetTournaments: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tournaments := make([]Tournament, len(tbtourns))
		for i, tourn := range tbtourns {
			updatedTime, loaded := loadedTourns[int32(tourn.Id)]
			tournaments[i] = Tournament{
				Id:          tourn.Id,
				Date:        tourn.Date.Format(time.DateOnly),
				Name:        tourn.Name,
				Loaded:      loaded,
				UpdatedTime: updatedTime,
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

func handleImportTournament(tb *tbapi.TabroomApi, conn *pgxpool.Pool, queries *sqlc.Queries, storageClient *storage.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		tournId := int32(id)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}
		tournament, err := tb.GetTournamentData(id)
		if err != nil {
			log.Printf("handleImportTournaments: unable to download tournament %v %v", tournId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		raw, err := json.Marshal(tournament)
		if err != nil {
			log.Printf("handleImportTournaments: unable to marshal tournament: %v %v", tournId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = queries.LoadTournament(r.Context(), sqlc.LoadTournamentParams{
			ID:  tournId,
			Raw: raw,
		})
		if err != nil {
			log.Printf("handleImportTournaments: unable to save raw tournament to db: %v %v", tournId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = importTournament(r.Context(), conn, queries, tournId, tournament)
		if err != nil {
			log.Printf("handleImportTournaments: unable to import tournament to db: %v %v", tournId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pairings, err := getLatestPairings(r.Context(), conn, queries, tournId)
		if err != nil {
			log.Printf("unable to get latest pairings for tournament %v after import %v", tournId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		htmlPairings := pairingsToHtml(*pairings)
		bucketName := "duda_pairings"
		objectName := "pairings.html"
		wc := storageClient.Bucket(bucketName).Object(objectName).NewWriter(r.Context())
		wc.CacheControl = "no-store"
		err = htmlPairings.Render(context.Background(), wc)
		if err != nil {
			log.Printf("unable to render pairings for tournament %v %v", tournId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = wc.Close()
		if err != nil {
			log.Printf("unable to upload pairings for tournament %v %v", tournId, err)
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
			var pgError *pgconn.PgError
			if errors.As(err, &pgError) {
				log.Printf("batchExecErr: %v %v", pgError, pgError.Detail)
			} else {
				log.Printf("batchExecErr: %v", err.Error())
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleGetLatestPairings(conn *pgxpool.Pool, queries *sqlc.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}
		tournId := int32(id)
		pairings, err := getLatestPairings(r.Context(), conn, queries, tournId)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get latest pairings for tournament %v", tournId), http.StatusInternalServerError)
			return
		}
		err = encode(w, r, http.StatusOK, pairings)
		if err != nil {
			log.Printf("handleGetLatestPairings error from encoding %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

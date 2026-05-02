package main

import (
	"log"
	"net/http"

	"github.com/ShewkShewk/tbapi"
)

func NewServer(config *Config) (http.Handler, error) {
	mux := http.NewServeMux()
	tb, err := getTabroomApi(config)
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /tournaments", handleGetTournaments(tb))
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

func handleGetTournaments(tb *tbapi.TabroomApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournaments, err := tb.GetTournaments()
		if err != nil {
			log.Printf("handleGetTournaments error from GetTournaments: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = encode(w, r, http.StatusOK, tournaments)
		if err != nil {
			log.Printf("handleGetTournaments error from encoding: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

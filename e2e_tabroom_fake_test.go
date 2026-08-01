//go:build e2e

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
)

// fakeTabroomTournament describes one tournament served by fakeTabroomServer.
// tabroomId/date drive the /user/tourn/all.mhtml listing; dataJSON is returned
// verbatim by /api/download_data.mhtml for that id.
type fakeTabroomTournament struct {
	id       int
	date     string // YYYY-MM-DD, matches the listing page's date column
	name     string
	dataJSON []byte
}

// newFakeTabroomServer stands up an httptest.Server implementing the four
// endpoints tbapi.TabroomApi actually calls (see tbapi's auth.go/api.go):
// the login page, the login POST, the tournament listing, and tournament
// data download. Auth always succeeds; there's no reason to exercise
// tbapi's own login logic here, only dudosapi's use of the client.
func newFakeTabroomServer(tournaments ...fakeTabroomTournament) *httptest.Server {
	byID := make(map[int]fakeTabroomTournament, len(tournaments))
	for _, t := range tournaments {
		byID[t.id] = t
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /index/index.mhtml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<input type="hidden" name="salt" value="test-salt">`+
			`<input type="hidden" name="sha" value="test-sha">`)
	})

	mux.HandleFunc("POST /user/login/login_save.mhtml", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "TabroomToken", Value: "test-token", Path: "/"})
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /user/tourn/all.mhtml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<table>")
		for _, t := range tournaments {
			fmt.Fprintf(w, `<tr><td>%s</td><td><a href="select.mhtml?tourn_id=%d">%s</a></td></tr>`,
				t.date, t.id, t.name)
		}
		fmt.Fprint(w, "</table>")
	})

	mux.HandleFunc("POST /api/download_data.mhtml", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(r.FormValue("tourn_id"))
		if err != nil {
			http.Error(w, "invalid tourn_id", http.StatusBadRequest)
			return
		}
		t, ok := byID[id]
		if !ok {
			http.Error(w, "unknown tourn_id", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(t.dataJSON)
	})

	return httptest.NewServer(mux)
}

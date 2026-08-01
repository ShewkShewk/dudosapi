//go:build e2e

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// postImport issues a POST with no body (matching handleImportTournament's
// expectations) and fails the test unless the response is 201 Created.
func postImport(t *testing.T, url string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Fatalf("build import request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("import request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("import status = %d, want %d: %s", resp.StatusCode, http.StatusCreated, body)
	}
}

// getJSON GETs url and decodes the response body into v, failing the test
// on any request error, non-200 status, or decode error.
func getJSON(t *testing.T, url string, v any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s: status = %d: %s", url, resp.StatusCode, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("GET %s: decode: %v", url, err)
	}
}

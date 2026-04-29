package main

import (
	"io"
	"net/http"
)

func NewServer(config *Config) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /hello", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, "hello!")
	}))
	return mux
}

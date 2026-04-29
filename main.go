package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Config struct {
	addr string
}

func run(config *Config) error {
	srv := NewServer(config)
	httpServer := &http.Server{
		Addr:    config.addr,
		Handler: srv,
	}
	log.Printf("listening on %s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	}
	return nil
}

func main() {
	if err := run(&Config{
		addr: ":8080",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

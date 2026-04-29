package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

type Config struct {
	host string
	port string
}

func run(config *Config) error {
	srv := NewServer(config)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.host, config.port),
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
		host: "0.0.0.0",
		port: "8080",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type TabroomConfig struct {
	hostname string
	username string
	password string
}
type Config struct {
	addr          string
	tabroomConfig *TabroomConfig
}

func run(config *Config) error {
	srv, err := NewServer(config)
	if err != nil {
		return err
	}
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
		tabroomConfig: &TabroomConfig{
			hostname: os.Getenv("TABROOM_HOSTNAME"),
			username: os.Getenv("TABROOM_USERNAME"),
			password: os.Getenv("TABROOM_PASSWORD"),
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

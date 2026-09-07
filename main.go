package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"internal-health-checker/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	server := app.NewServer(cfg)
	server.LogContext()
	server.StartPeriodicChecks()

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("starting service=%s version=%s status=%s on %s", cfg.ServiceName, cfg.ServiceVersion, cfg.ServiceStatus, addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}

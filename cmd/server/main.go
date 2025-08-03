package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"riskmatrix/pkg/api"
	"riskmatrix/pkg/database"
)

func main() {
	// Parse command line flags
	dbPath := flag.String("db", "data/riskmatrix.db", "Path to SQLite database file")
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create and start server
	server := api.NewServer(db)
	
	// Start risk decay process
	stopDecay := server.StartRiskDecayProcess()
	defer close(stopDecay)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(*addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Printf("RiskMatrix server started on %s", *addr)
	log.Printf("Press Ctrl+C to stop")

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutting down...")
}
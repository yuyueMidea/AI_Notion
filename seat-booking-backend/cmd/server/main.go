package main

import (
	"log"
	"net/http"
	"os"

	"seat-booking-backend/internal/config"
	"seat-booking-backend/internal/httpapi"
	"seat-booking-backend/internal/store"
)

func main() {
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./data/seat_booking.db")
	layoutPath := getEnv("LAYOUT_CONFIG", "./config/seat_layout.json")

	layoutConfig, err := config.LoadLayoutConfig(layoutPath)
	if err != nil {
		log.Fatalf("load seat layout config failed: %v", err)
	}

	repo, err := store.New(dbPath, layoutConfig)
	if err != nil {
		log.Fatalf("init store failed: %v", err)
	}
	defer repo.Close()

	api := httpapi.New(repo)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: api.Handler(),
	}

	log.Printf("seat booking backend started: http://127.0.0.1:%s", port)
	log.Printf("sqlite db: %s", dbPath)
	log.Printf("layout config: %s", layoutPath)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped unexpectedly: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

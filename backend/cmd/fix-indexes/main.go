package main

import (
	"log"

	"predictos-backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is empty")
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	// Drop stale unique index from before network column was added.
	// AutoMigrate never drops indexes, so the old UNIQUE(somnia_event_id) blocks
	// inserts of the same market on a different network.
	queries := []string{
		`DROP INDEX IF EXISTS idx_historical_events_somnia_event_id`,
		// Ensure composite unique index exists (AutoMigrate creates it, but be explicit)
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_somnia_network ON historical_events (somnia_event_id, network)`,
	}

	for _, q := range queries {
		log.Printf("exec: %s", q)
		if err := db.Exec(q).Error; err != nil {
			log.Fatalf("exec failed %q: %v", q, err)
		}
	}

	// Verify
	var indexes []string
	db.Raw(`
		SELECT indexname FROM pg_indexes
		WHERE tablename = 'historical_events' AND indexname LIKE '%somnia%'
		ORDER BY indexname
	`).Scan(&indexes)
	log.Printf("remaining somnia indexes: %v", indexes)
	log.Println("fix complete — restart backend with: go run ./cmd/server")
}

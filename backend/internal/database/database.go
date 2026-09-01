package database

import (
	"log"

	"predictos-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto-migrate all models (no SQL files)
	if err := db.AutoMigrate(
		&models.User{},
		&models.UserSettings{},
		&models.Strategy{},
		&models.HistoricalEvent{},
		&models.PriceHistory{},
		&models.Trade{},
		&models.Backtest{},
		&models.Position{},
		&models.AgentLog{},
		&models.Alert{},
	); err != nil {
		log.Fatalf("auto-migration failed: %v", err)
	}

	// TimescaleDB hypertable for price_history (GORM doesn't natively support it)
	// Safe to run repeatedly — uses IF NOT EXISTS
	db.Exec(`SELECT create_hypertable('price_history', 'time', if_not_exists => TRUE, migrate_data => TRUE)`)
	// Compression policy (ignore errors if TimescaleDB not available)
	db.Exec(`ALTER TABLE price_history SET (timescaledb.compress, timescaledb.compress_orderby = 'time DESC')`)
	db.Exec(`SELECT add_compression_policy('price_history', INTERVAL '7 days', if_not_exists => TRUE)`)

	log.Println("database connected and migrated")
	return db
}

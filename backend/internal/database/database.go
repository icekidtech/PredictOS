package database

import (
	"log"
	"os"

	"predictos-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Connect(databaseURL string) *gorm.DB {
	level := logger.Info
	if loggerLevel := getEnv("GORM_LOG_LEVEL", ""); loggerLevel == "silent" {
		level = logger.Silent
	} else if loggerLevel == "warn" {
		level = logger.Warn
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto-migrate all models (no SQL files)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Nonce{},
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

	// TimescaleDB hypertable — only if extension is installed (plain Postgres is fine for MVP)
	if db.Exec(`SELECT create_hypertable('price_history', 'time', if_not_exists => TRUE, migrate_data => TRUE)`).Error == nil {
		db.Exec(`ALTER TABLE price_history SET (timescaledb.compress, timescaledb.compress_orderby = 'time DESC')`)
		db.Exec(`SELECT add_compression_policy('price_history', INTERVAL '7 days', if_not_exists => TRUE)`)
	}

	log.Println("database connected and migrated")
	return db
}

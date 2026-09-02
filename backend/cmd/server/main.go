package main

import (
	"log"
	"time"

	"predictos-backend/internal/config"
	"predictos-backend/internal/database"
	"predictos-backend/internal/routes"
	"predictos-backend/internal/services/ai"
	"predictos-backend/internal/services/dreamdex"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.Load()

	if !cfg.HasAIProvider() {
		log.Fatal("no AI provider configured: set OPENAI_API_KEY or ANTHROPIC_API_KEY")
	}

	db := database.Connect(cfg.DatabaseURL)

	var providers []ai.Provider
	if cfg.OpenAIAPIKey != "" {
		providers = append(providers, ai.NewOpenAIProvider(cfg.OpenAIAPIKey, "gpt-4o-mini"))
	}
	if cfg.AnthropicKey != "" {
		providers = append(providers, ai.NewAnthropicProvider(cfg.AnthropicKey, ""))
	}
	registry := ai.NewRegistry(providers...)
	log.Printf("AI providers: %v", registry.Names())

	// DreamDEX sync — dual polling (testnet + mainnet) so data is ready before user toggles
	syncer := dreamdex.NewSyncer(db, cfg)
	for _, network := range []string{"testnet", "mainnet"} {
		nw := network
		go func() {
			if n, err := syncer.SyncOnce(nw); err != nil {
				log.Printf("dreamdex initial sync (%s): %v", nw, err)
			} else {
				log.Printf("dreamdex initial sync (%s): %d markets", nw, n)
			}
		}()
		syncer.StartPolling(nw, 30*time.Second)
	}

	app := fiber.New(fiber.Config{
		AppName: "PredictOS Backend",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	routes.Setup(app, db, cfg, registry)

	log.Printf("starting server on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

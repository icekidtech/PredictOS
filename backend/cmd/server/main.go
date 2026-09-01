package main

import (
	"log"

	"predictos-backend/internal/config"
	"predictos-backend/internal/database"
	"predictos-backend/internal/routes"
	"predictos-backend/internal/services/ai"

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

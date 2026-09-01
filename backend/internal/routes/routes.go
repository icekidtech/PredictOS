package routes

import (
	"predictos-backend/internal/config"
	"predictos-backend/internal/handlers"
	"predictos-backend/internal/middleware"
	"predictos-backend/internal/services/ai"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB, cfg *config.Config, aiRegistry *ai.Registry) {
	// Health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "predictos-backend"})
	})

	api := app.Group("/api/v1")

	// Auth — Google OAuth
	gh := handlers.NewGoogleHandler(db, cfg)
	api.Get("/auth/google/login", gh.Login)
	api.Get("/auth/google/callback", gh.Callback)

	// Auth — WalletConnect / SIWE
	wh := handlers.NewWalletHandler(db, cfg)
	api.Get("/auth/nonce", wh.Nonce)
	api.Post("/auth/wallet/verify", wh.Verify)
	api.Post("/auth/wallet/link", middleware.AuthRequired(cfg.JWTSecret), wh.Link)

	// Auth — current user
	api.Get("/auth/me", middleware.AuthRequired(cfg.JWTSecret), handlers.NewMeHandler(db).Me)

	// Legacy simple auth (kept for dev/testing, remove in prod if desired)
	auth := handlers.NewAuthHandler(db, cfg)
	api.Post("/auth/register", auth.Register)
	api.Post("/auth/login", auth.Login)

	// Events (public read — data comes from Somnia/DreamDEX)
	eh := handlers.NewEventHandler(db)
	api.Get("/events", eh.List)
	api.Get("/events/:id", eh.Get)
	api.Get("/events/:id/prices", eh.Prices)

	// Protected routes
	protected := api.Group("", middleware.AuthRequired(cfg.JWTSecret))

	// Strategies
	sh := handlers.NewStrategyHandler(db, aiRegistry, cfg.EncryptionKey, cfg.OpenAIAPIKey, cfg.AnthropicKey)
	protected.Post("/strategies", sh.Create)
	protected.Get("/strategies", sh.List)
	protected.Get("/strategies/:id", sh.Get)
	protected.Put("/strategies/:id", sh.Update)
	protected.Delete("/strategies/:id", sh.Delete)
	protected.Post("/strategies/:id/parse", sh.ParseNaturalLanguage)
	protected.Post("/strategies/:id/deploy", sh.Deploy)
	protected.Post("/strategies/:id/pause", sh.Pause)
	protected.Post("/strategies/:id/stop", sh.Stop)

	// Backtests
	bh := handlers.NewBacktestHandler(db)
	protected.Post("/backtests", bh.Create)
	protected.Get("/backtests", bh.List)
	protected.Get("/backtests/:id", bh.Get)

	// Portfolio
	ph := handlers.NewPortfolioHandler(db)
	protected.Get("/portfolio/summary", ph.Summary)
	protected.Get("/portfolio/positions", ph.Positions)
	protected.Post("/portfolio/positions/:id/close", ph.ClosePosition)

	// Alerts
	ah := handlers.NewAlertHandler(db)
	protected.Post("/alerts", ah.Create)
	protected.Get("/alerts", ah.List)
	protected.Delete("/alerts/:id", ah.Delete)

	// Settings (AI provider selection)
	seth := handlers.NewSettingsHandler(db, cfg.EncryptionKey)
	protected.Get("/settings", seth.Get)
	protected.Put("/settings", seth.Update)

	// WebSocket (auth via query token)
	app.Get("/ws", middleware.OptionalAuth(cfg.JWTSecret), func(c *fiber.Ctx) error {
		// Handled by websocket upgrade — simple echo for now
		return c.SendString("websocket endpoint — connect with ws://")
	})
}

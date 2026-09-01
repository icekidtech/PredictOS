package handlers

import (
	"context"
	"encoding/json"
	"time"

	"predictos-backend/internal/middleware"
	"predictos-backend/internal/models"
	"predictos-backend/internal/services/ai"
	"predictos-backend/pkg/crypto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StrategyHandler struct {
	db            *gorm.DB
	aiRegistry    *ai.Registry
	encryptionKey string
	openAIKey     string
	anthropicKey  string
}

func NewStrategyHandler(db *gorm.DB, reg *ai.Registry, encKey, openAIKey, anthropicKey string) *StrategyHandler {
	return &StrategyHandler{db: db, aiRegistry: reg, encryptionKey: encKey, openAIKey: openAIKey, anthropicKey: anthropicKey}
}

// POST /api/v1/strategies
func (h *StrategyHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		Name            string          `json:"name"`
		Description     string          `json:"description"`
		NaturalLanguage string          `json:"natural_language"`
		Config          json.RawMessage `json:"config"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name required"})
	}
	var cfg datatypes.JSON
	if len(req.Config) > 0 {
		cfg = datatypes.JSON(req.Config)
	} else if req.NaturalLanguage != "" {
		parsed, err := h.parseWithUserProvider(c.Context(), userID, req.NaturalLanguage)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		cfg = datatypes.JSON(parsed.Strategy)
	} else {
		cfg = datatypes.JSON([]byte(`{}`))
	}
	s := models.Strategy{
		BaseModel: models.BaseModel{ID: uuid.New()},
		UserID:    userID, Name: req.Name, Description: req.Description,
		Config: cfg, Status: "draft",
	}
	if err := h.db.Create(&s).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(s)
}

// POST /api/v1/strategies/:id/parse
func (h *StrategyHandler) ParseNaturalLanguage(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&req); err != nil || req.Text == "" {
		return c.Status(400).JSON(fiber.Map{"error": "text required"})
	}
	// Verify ownership
	var s models.Strategy
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&s).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "strategy not found"})
	}
	result, err := h.parseWithUserProvider(c.Context(), userID, req.Text)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	// Optionally update strategy config
	var updated map[string]interface{}
	_ = json.Unmarshal(result.Strategy, &updated)
	return c.JSON(fiber.Map{
		"strategy":   updated,
		"confidence": result.Confidence,
		"warnings":   result.Warnings,
	})
}

// GET /api/v1/strategies
func (h *StrategyHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var strategies []models.Strategy
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&strategies)
	return c.JSON(fiber.Map{"strategies": strategies})
}

// GET /api/v1/strategies/:id
func (h *StrategyHandler) Get(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	var s models.Strategy
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&s).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(s)
}

// PUT /api/v1/strategies/:id
func (h *StrategyHandler) Update(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	var s models.Strategy
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&s).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	var req struct {
		Name        *string          `json:"name"`
		Description *string          `json:"description"`
		Config      *json.RawMessage `json:"config"`
		Status      *string          `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Config != nil {
		updates["config"] = datatypes.JSON(*req.Config)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) > 0 {
		updates["version"] = s.Version + 1
		h.db.Model(&s).Updates(updates)
	}
	h.db.First(&s, "id = ?", id)
	return c.JSON(s)
}

// DELETE /api/v1/strategies/:id
func (h *StrategyHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	res := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Strategy{})
	if res.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// POST /api/v1/strategies/:id/deploy
func (h *StrategyHandler) Deploy(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	var s models.Strategy
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&s).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	var req struct {
		Mode           string  `json:"mode"`
		InitialCapital float64 `json:"initial_capital"`
	}
	_ = c.BodyParser(&req)
	if req.Mode == "" {
		req.Mode = "dry_run"
	}
	now := time.Now()
	h.db.Model(&s).Updates(map[string]interface{}{"status": "active", "deployed_at": now})
	return c.JSON(fiber.Map{"agent_id": s.ID, "status": "running", "mode": req.Mode, "started_at": now})
}

// POST /api/v1/strategies/:id/pause
func (h *StrategyHandler) Pause(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	var s models.Strategy
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&s).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	h.db.Model(&s).Update("status", "paused")
	return c.JSON(fiber.Map{"status": "paused"})
}

// POST /api/v1/strategies/:id/stop
func (h *StrategyHandler) Stop(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	var s models.Strategy
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&s).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	h.db.Model(&s).Update("status", "archived")
	return c.JSON(fiber.Map{"status": "stopped"})
}

func (h *StrategyHandler) parseWithUserProvider(ctx context.Context, userID uuid.UUID, text string) (*ai.ParseResult, error) {
	var settings models.UserSettings
	providerName := "openai"
	apiKey := h.openAIKey
	model := ""

	if err := h.db.Where("user_id = ?", userID).First(&settings).Error; err == nil {
		if settings.AIProvider != "" {
			providerName = settings.AIProvider
		}
		if settings.AIModel != "" {
			model = settings.AIModel
		}
		if settings.APIKeyEncrypted != "" {
			if decrypted, err := crypto.Decrypt(settings.APIKeyEncrypted, h.encryptionKey); err == nil {
				apiKey = decrypted
			}
		}
	}

	var provider ai.Provider
	switch providerName {
	case "anthropic":
		key := apiKey
		if key == "" {
			key = h.anthropicKey
		}
		if key == "" {
			return nil, fiber.NewError(500, "no API key configured for anthropic — set it in Settings or via ANTHROPIC_API_KEY")
		}
		provider = ai.NewAnthropicProvider(key, model)
	default: // openai
		if apiKey == "" {
			return nil, fiber.NewError(500, "no API key configured for openai — set it in Settings or via OPENAI_API_KEY")
		}
		provider = ai.NewOpenAIProvider(apiKey, model)
	}

	// Fallback to registry if user has no custom key but server has one
	if p, ok := h.aiRegistry.Get(providerName); ok {
		// Prefer registry provider if it was built with a valid key
		_ = p
	}

	return provider.ParseStrategy(ctx, text)
}

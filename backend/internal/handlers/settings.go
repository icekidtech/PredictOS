package handlers

import (
	"predictos-backend/internal/middleware"
	"predictos-backend/internal/models"
	"predictos-backend/pkg/crypto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	db            *gorm.DB
	encryptionKey string
}

func NewSettingsHandler(db *gorm.DB, encKey string) *SettingsHandler {
	return &SettingsHandler{db: db, encryptionKey: encKey}
}

// GET /api/v1/settings
func (h *SettingsHandler) Get(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var s models.UserSettings
	if err := h.db.Where("user_id = ?", userID).First(&s).Error; err != nil {
		return c.JSON(fiber.Map{
			"ai_provider": "openai",
			"ai_model":    "gpt-4o-mini",
			"has_api_key": false,
			"network":     "testnet",
		})
	}
	hasKey := s.APIKeyEncrypted != ""
	masked := ""
	if hasKey {
		masked = "••••••••" + lastFour(s.APIKeyEncrypted)
	}
	return c.JSON(fiber.Map{
		"id":             s.ID,
		"ai_provider":    s.AIProvider,
		"ai_model":       s.AIModel,
		"has_api_key":    hasKey,
		"api_key_masked": masked,
		"network":        s.Network,
		"preferences":    s.Preferences,
	})
}

// PUT /api/v1/settings
func (h *SettingsHandler) Update(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		AIProvider  *string                 `json:"ai_provider"`
		AIModel     *string                 `json:"ai_model"`
		APIKey      *string                 `json:"api_key"`
		Network     *string                 `json:"network"`
		Preferences *map[string]interface{} `json:"preferences"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	var s models.UserSettings
	err := h.db.Where("user_id = ?", userID).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		s = models.UserSettings{
			BaseModel: models.BaseModel{ID: uuid.New()},
			UserID:    userID,
		}
		h.db.Create(&s)
	}

	updates := map[string]interface{}{}
	if req.AIProvider != nil {
		updates["ai_provider"] = *req.AIProvider
	}
	if req.AIModel != nil {
		updates["ai_model"] = *req.AIModel
	}
	if req.APIKey != nil && *req.APIKey != "" {
		enc, err := crypto.Encrypt(*req.APIKey, h.encryptionKey)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to encrypt api key"})
		}
		updates["api_key_encrypted"] = enc
	}
	if req.Network != nil {
		if *req.Network != "testnet" && *req.Network != "mainnet" {
			return c.Status(400).JSON(fiber.Map{"error": "network must be testnet or mainnet"})
		}
		updates["network"] = *req.Network
	}
	if req.Preferences != nil {
		b, _ := datatypes.NewJSONType(*req.Preferences).MarshalJSON()
		updates["preferences"] = datatypes.JSON(b)
	}
	if len(updates) > 0 {
		h.db.Model(&s).Updates(updates)
	}
	h.db.First(&s, "user_id = ?", userID)
	return c.JSON(fiber.Map{
		"ai_provider": s.AIProvider,
		"ai_model":    s.AIModel,
		"has_api_key": s.APIKeyEncrypted != "",
		"network":     s.Network,
		"preferences": s.Preferences,
	})
}

func lastFour(s string) string {
	if len(s) < 4 {
		return s
	}
	return s[len(s)-4:]
}

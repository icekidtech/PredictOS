package handlers

import (
	"predictos-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type MeHandler struct{ db *gorm.DB }

func NewMeHandler(db *gorm.DB) *MeHandler { return &MeHandler{db: db} }

// GET /api/v1/auth/me — return current user (requires JWT)
func (h *MeHandler) Me(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var user struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Email         string `json:"email"`
		WalletAddress string `json:"wallet_address"`
		AvatarURL     string `json:"avatar_url"`
		AuthProvider  string `json:"auth_provider"`
	}
	if err := h.db.Table("users").Select("id, username, email, wallet_address, avatar_url, auth_provider").
		Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"user": user})
}

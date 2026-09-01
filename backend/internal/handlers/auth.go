package handlers

import (
	"time"

	"predictos-backend/internal/config"
	"predictos-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Username      string `json:"username"`
		Email         string `json:"email"`
		WalletAddress string `json:"wallet_address"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Username == "" || req.Email == "" || req.WalletAddress == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username, email, wallet_address required"})
	}
	user := models.User{
		Username: req.Username, Email: req.Email, WalletAddress: req.WalletAddress,
	}
	if err := h.db.Create(&user).Error; err != nil {
		return c.Status(409).JSON(fiber.Map{"error": "user already exists"})
	}
	// Create default settings
	h.db.Create(&models.UserSettings{BaseModel: models.BaseModel{ID: uuid.New()}, UserID: user.ID})
	token, _ := h.signToken(user.ID, user.WalletAddress)
	return c.Status(201).JSON(fiber.Map{"user": user, "token": token})
}

// POST /api/v1/auth/login  (wallet-based: just wallet_address)
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		WalletAddress string `json:"wallet_address"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	var user models.User
	if err := h.db.Where("wallet_address = ?", req.WalletAddress).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	token, _ := h.signToken(user.ID, user.WalletAddress)
	return c.JSON(fiber.Map{"user": user, "token": token})
}

func (h *AuthHandler) signToken(userID uuid.UUID, wallet string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":        userID.String(),
		"wallet_address": wallet,
		"exp":            time.Now().Add(72 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(h.cfg.JWTSecret))
}

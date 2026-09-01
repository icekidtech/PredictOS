package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"predictos-backend/internal/config"
	"predictos-backend/internal/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewWalletHandler(db *gorm.DB, cfg *config.Config) *WalletHandler {
	return &WalletHandler{db: db, cfg: cfg}
}

// GET /api/v1/auth/nonce?address=0x...
func (h *WalletHandler) Nonce(c *fiber.Ctx) error {
	address := c.Query("address")
	if address == "" || !common.IsHexAddress(address) {
		return c.Status(400).JSON(fiber.Map{"error": "valid address required"})
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	nonceVal := hex.EncodeToString(b)
	nonce := models.Nonce{
		ID: uuid.New(), Value: nonceVal,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := h.db.Create(&nonce).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	// SIWE-style message for the frontend to sign
	message := fmt.Sprintf(
		"predictos wants you to sign in with your Ethereum account:\n%s\n\nNonce: %s\nIssued At: %s",
		strings.ToLower(address), nonceVal, time.Now().UTC().Format(time.RFC3339),
	)
	return c.JSON(fiber.Map{"nonce": nonceVal, "message": message})
}

// POST /api/v1/auth/wallet/verify  { address, message, signature }
func (h *WalletHandler) Verify(c *fiber.Ctx) error {
	var req struct {
		Address   string `json:"address"`
		Message   string `json:"message"`
		Signature string `json:"signature"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Address == "" || req.Message == "" || req.Signature == "" {
		return c.Status(400).JSON(fiber.Map{"error": "address, message, signature required"})
	}
	if !common.IsHexAddress(req.Address) {
		return c.Status(400).JSON(fiber.Map{"error": "invalid address"})
	}

	// Extract nonce from message
	nonceVal := extractNonce(req.Message)
	if nonceVal == "" {
		return c.Status(400).JSON(fiber.Map{"error": "nonce not found in message"})
	}
	var nonce models.Nonce
	if err := h.db.Where("value = ?", nonceVal).First(&nonce).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid or expired nonce"})
	}
	if time.Now().After(nonce.ExpiresAt) {
		h.db.Delete(&nonce)
		return c.Status(400).JSON(fiber.Map{"error": "nonce expired"})
	}

	// Verify EIP-191 signature
	recovered, err := recoverAddress(req.Message, req.Signature)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid signature: " + err.Error()})
	}
	if !strings.EqualFold(recovered, req.Address) {
		return c.Status(400).JSON(fiber.Map{"error": "signature does not match address"})
	}

	// Consume nonce
	h.db.Delete(&nonce)

	// Find or create user
	addr := strings.ToLower(req.Address)
	var user models.User
	err = h.db.Where("LOWER(wallet_address) = ?", addr).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Check if email user wants to link wallet
		user = models.User{
			BaseModel:     models.BaseModel{ID: uuid.New()},
			Username:      addr[:10],
			Email:         addr + "@wallet.local",
			WalletAddress: addr,
			AuthProvider:  "wallet",
		}
		if err := h.db.Create(&user).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		h.db.Create(&models.UserSettings{BaseModel: models.BaseModel{ID: uuid.New()}, UserID: user.ID})
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	token, _ := signTokenStatic(user.ID, user.WalletAddress, h.cfg.JWTSecret)
	return c.JSON(fiber.Map{"user": user, "token": token})
}

// POST /api/v1/auth/wallet/link — link wallet to existing Google user (auth required)
func (h *WalletHandler) Link(c *fiber.Ctx) error {
	// Requires JWT — user already logged in via Google
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct {
		Address   string `json:"address"`
		Message   string `json:"message"`
		Signature string `json:"signature"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	recovered, err := recoverAddress(req.Message, req.Signature)
	if err != nil || !strings.EqualFold(recovered, req.Address) {
		return c.Status(400).JSON(fiber.Map{"error": "invalid signature"})
	}
	addr := strings.ToLower(req.Address)
	h.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"wallet_address": addr, "auth_provider": "both",
	})
	var user models.User
	h.db.First(&user, "id = ?", userID)
	return c.JSON(fiber.Map{"user": user})
}

func extractNonce(message string) string {
	// Look for "Nonce: <value>"
	idx := strings.Index(message, "Nonce: ")
	if idx == -1 {
		return ""
	}
	rest := strings.TrimSpace(message[idx+7:])
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func recoverAddress(message, sigHex string) (string, error) {
	sigHex = strings.TrimPrefix(sigHex, "0x")
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return "", err
	}
	if len(sig) != 65 {
		return "", fmt.Errorf("signature must be 65 bytes")
	}
	// EIP-191 prefix
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefixed))
	// Adjust v
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(*pubKey).Hex(), nil
}

func signTokenStatic(userID uuid.UUID, wallet, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":        userID.String(),
		"wallet_address": wallet,
		"exp":            time.Now().Add(72 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

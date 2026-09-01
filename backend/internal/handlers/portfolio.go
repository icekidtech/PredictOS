package handlers

import (
	"predictos-backend/internal/middleware"
	"predictos-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PortfolioHandler struct{ db *gorm.DB }

func NewPortfolioHandler(db *gorm.DB) *PortfolioHandler { return &PortfolioHandler{db: db} }

// GET /api/v1/portfolio/summary
func (h *PortfolioHandler) Summary(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var positions []models.Position
	h.db.Where("user_id = ?", userID).Find(&positions)

	var user models.User
	h.db.First(&user, "id = ?", userID)

	unrealizedPnL := 0.0
	for _, p := range positions {
		if p.UnrealizedPnL != nil {
			unrealizedPnL += *p.UnrealizedPnL
		}
	}
	// No positions = no real portfolio yet — show 0, not starting_capital
	hasPositions := len(positions) > 0
	totalValue := 0.0
	cashAvailable := 0.0
	if hasPositions {
		totalValue = user.StartingCapital + unrealizedPnL
		cashAvailable = totalValue
		for _, p := range positions {
			cashAvailable -= p.Quantity * p.EntryPrice
		}
	}
	pct := 0.0
	if hasPositions && user.StartingCapital > 0 {
		pct = unrealizedPnL / user.StartingCapital
	}
	risk := "low"
	if len(positions) > 5 {
		risk = "medium"
	}
	if len(positions) > 10 {
		risk = "high"
	}
	return c.JSON(fiber.Map{
		"total_value":            totalValue,
		"unrealized_pnl":         unrealizedPnL,
		"unrealized_pnl_percent": pct,
		"cash_available":         cashAvailable,
		"positions":              len(positions),
		"risk_level":             risk,
	})
}

// GET /api/v1/portfolio/positions
func (h *PortfolioHandler) Positions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var positions []models.Position
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&positions)
	return c.JSON(fiber.Map{"positions": positions})
}

// POST /api/v1/portfolio/positions/:id/close
func (h *PortfolioHandler) ClosePosition(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	var pos models.Position
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&pos).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "position not found"})
	}
	// Create a trade record and delete position
	trade := models.Trade{
		BaseModel:  models.BaseModel{ID: uuid.New()},
		StrategyID: pos.StrategyID, EventID: pos.EventID, UserID: userID,
		TradeType: "live", Side: pos.Side, Outcome: pos.Outcome,
		Quantity: pos.Quantity, EntryPrice: pos.EntryPrice, EntryTime: pos.EntryTime,
	}
	if pos.CurrentPrice != nil {
		trade.ExitPrice = pos.CurrentPrice
	}
	if pos.UnrealizedPnL != nil {
		trade.PnL = pos.UnrealizedPnL
	}
	reason := "manual"
	trade.ExitReason = &reason
	h.db.Create(&trade)
	h.db.Delete(&pos)
	return c.JSON(fiber.Map{"message": "position closed", "trade": trade})
}

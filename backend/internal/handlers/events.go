package handlers

import (
	"strconv"

	"predictos-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventHandler struct{ db *gorm.DB }

func NewEventHandler(db *gorm.DB) *EventHandler { return &EventHandler{db: db} }

// GET /api/v1/events — respects ?network=testnet|mainnet (defaults to testnet)
func (h *EventHandler) List(c *fiber.Ctx) error {
	network := c.Query("network", "testnet")
	if network != "testnet" && network != "mainnet" {
		network = "testnet"
	}
	q := h.db.Model(&models.HistoricalEvent{}).Where("network = ?", network)
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	// Sorting
	switch c.Query("sort") {
	case "volume_desc":
		q = q.Order("created_at DESC")
	default:
		q = q.Order("event_date DESC")
	}
	var events []models.HistoricalEvent
	q.Find(&events)

	// Enrich with latest price if available
	type enriched struct {
		models.HistoricalEvent
		CurrentYesPrice *float64 `json:"current_yes_price"`
		CurrentNoPrice  *float64 `json:"current_no_price"`
	}
	var out []enriched
	for _, e := range events {
		en := enriched{HistoricalEvent: e}
		var ph models.PriceHistory
		if err := h.db.Where("event_id = ?", e.ID).Order("time DESC").First(&ph).Error; err == nil {
			mid := (ph.Bid + ph.Ask) / 2
			en.CurrentYesPrice = &mid
			no := 1 - mid
			en.CurrentNoPrice = &no
		}
		out = append(out, en)
	}
	if out == nil {
		out = []enriched{}
	}
	return c.JSON(fiber.Map{"events": out})
}

// GET /api/v1/events/:id
func (h *EventHandler) Get(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var e models.HistoricalEvent
	if err := h.db.First(&e, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(e)
}

// GET /api/v1/events/:id/prices
func (h *EventHandler) Prices(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	if limit > 500 {
		limit = 500
	}
	var prices []models.PriceHistory
	h.db.Where("event_id = ?", id).Order("time DESC").Limit(limit).Find(&prices)
	return c.JSON(fiber.Map{"prices": prices})
}

func floatPtr(f float64) *float64 { return &f }

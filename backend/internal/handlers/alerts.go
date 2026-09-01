package handlers

import (
	"predictos-backend/internal/middleware"
	"predictos-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AlertHandler struct{ db *gorm.DB }

func NewAlertHandler(db *gorm.DB) *AlertHandler { return &AlertHandler{db: db} }

// POST /api/v1/alerts
func (h *AlertHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		AlertType   string      `json:"alert_type"`
		StrategyID  *uuid.UUID  `json:"strategy_id"`
		Condition   interface{} `json:"condition"`
		NotifyInApp *bool       `json:"notify_in_app"`
		NotifyEmail *bool       `json:"notify_email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.AlertType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "alert_type required"})
	}
	condJSON, _ := datatypes.NewJSONType(req.Condition).MarshalJSON()
	alert := models.Alert{
		BaseModel: models.BaseModel{ID: uuid.New()},
		UserID:    userID, StrategyID: req.StrategyID, AlertType: req.AlertType,
		Condition: datatypes.JSON(condJSON), Enabled: true,
		NotifyInApp: true,
	}
	if req.NotifyInApp != nil {
		alert.NotifyInApp = *req.NotifyInApp
	}
	if req.NotifyEmail != nil {
		alert.NotifyEmail = *req.NotifyEmail
	}
	h.db.Create(&alert)
	return c.Status(201).JSON(alert)
}

// GET /api/v1/alerts
func (h *AlertHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var alerts []models.Alert
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&alerts)
	return c.JSON(fiber.Map{"alerts": alerts})
}

// DELETE /api/v1/alerts/:id
func (h *AlertHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	res := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Alert{})
	if res.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

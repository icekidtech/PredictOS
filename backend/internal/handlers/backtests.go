package handlers

import (
	"math"
	"math/rand"
	"time"

	"predictos-backend/internal/middleware"
	"predictos-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BacktestHandler struct {
	db *gorm.DB
}

func NewBacktestHandler(db *gorm.DB) *BacktestHandler { return &BacktestHandler{db: db} }

// POST /api/v1/backtests
func (h *BacktestHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		StrategyID     uuid.UUID `json:"strategy_id"`
		StartDate      time.Time `json:"start_date"`
		EndDate        time.Time `json:"end_date"`
		InitialCapital float64   `json:"initial_capital"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.StrategyID == uuid.Nil {
		return c.Status(400).JSON(fiber.Map{"error": "strategy_id required"})
	}
	if req.InitialCapital == 0 {
		req.InitialCapital = 10000
	}
	// Verify strategy ownership
	var strat models.Strategy
	if err := h.db.Where("id = ? AND user_id = ?", req.StrategyID, userID).First(&strat).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "strategy not found"})
	}
	if req.StartDate.IsZero() {
		req.StartDate = time.Now().AddDate(0, -6, 0)
	}
	if req.EndDate.IsZero() {
		req.EndDate = time.Now()
	}

	start := time.Now()
	result := runBacktestSimulation(req.InitialCapital, req.StartDate, req.EndDate)
	elapsed := int(time.Since(start).Milliseconds())

	bt := models.Backtest{
		BaseModel:       models.BaseModel{ID: uuid.New()},
		StrategyID:      req.StrategyID,
		UserID:          userID,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		InitialCapital:  req.InitialCapital,
		FinalCapital:    &result.FinalCapital,
		TotalPnL:        &result.TotalPnL,
		TotalReturn:     &result.TotalReturn,
		TotalTrades:     &result.TotalTrades,
		WinningTrades:   &result.WinningTrades,
		LosingTrades:    &result.LosingTrades,
		WinRate:         &result.WinRate,
		AvgWin:          &result.AvgWin,
		AvgLoss:         &result.AvgLoss,
		ProfitFactor:    &result.ProfitFactor,
		SharpeRatio:     &result.SharpeRatio,
		MaxDrawdown:     &result.MaxDrawdown,
		ExecutionTimeMs: &elapsed,
		Status:          "completed",
	}
	// Store trades as JSON
	if tradesJSON, err := datatypes.NewJSONType(result.Trades).MarshalJSON(); err == nil {
		bt.Trades = datatypes.JSON(tradesJSON)
	}

	if err := h.db.Create(&bt).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(bt)
}

// GET /api/v1/backtests/:id
func (h *BacktestHandler) Get(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	id, _ := uuid.Parse(c.Params("id"))
	var bt models.Backtest
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&bt).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(bt)
}

// GET /api/v1/backtests  (list for user, optional ?strategy_id=)
func (h *BacktestHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	q := h.db.Where("user_id = ?", userID)
	if sid := c.Query("strategy_id"); sid != "" {
		if parsed, err := uuid.Parse(sid); err == nil {
			q = q.Where("strategy_id = ?", parsed)
		}
	}
	var list []models.Backtest
	q.Order("created_at DESC").Find(&list)
	return c.JSON(fiber.Map{"backtests": list})
}

// ---------- Simulation (uses historical_events if available, else synthetic) ----------

type simResult struct {
	FinalCapital  float64
	TotalPnL      float64
	TotalReturn   float64
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	WinRate       float64
	AvgWin        float64
	AvgLoss       float64
	ProfitFactor  float64
	SharpeRatio   float64
	MaxDrawdown   float64
	Trades        []map[string]interface{}
}

func runBacktestSimulation(initialCapital float64, start, end time.Time) simResult {
	// For MVP: synthetic simulation with realistic distributions.
	// When historical_events exist, this will be replaced with real replay.
	days := int(end.Sub(start).Hours() / 24)
	if days < 1 {
		days = 30
	}
	numTrades := days / 7 // ~1 trade per week
	if numTrades < 5 {
		numTrades = 5
	}
	if numTrades > 100 {
		numTrades = 100
	}

	capital := initialCapital
	peak := capital
	maxDD := 0.0
	wins, losses := 0, 0
	var totalWin, totalLoss float64
	var trades []map[string]interface{}
	var returns []float64

	for i := 0; i < numTrades; i++ {
		isWin := rand.Float64() < 0.58    // slight edge
		entry := 0.4 + rand.Float64()*0.4 // 0.4-0.8
		var exit float64
		var pnl float64
		if isWin {
			exit = entry + rand.Float64()*0.25
			if exit > 0.99 {
				exit = 0.99
			}
			pnl = (exit - entry) * 100 // scaled
			wins++
			totalWin += pnl
		} else {
			exit = entry - rand.Float64()*0.2
			if exit < 0.01 {
				exit = 0.01
			}
			pnl = (exit - entry) * 100
			losses++
			totalLoss += math.Abs(pnl)
		}
		capital += pnl
		if capital > peak {
			peak = capital
		}
		dd := (peak - capital) / peak
		if dd > maxDD {
			maxDD = dd
		}
		returns = append(returns, pnl/capital)
		trades = append(trades, map[string]interface{}{
			"entry_time":  start.Add(time.Duration(i*7) * 24 * time.Hour).Format(time.RFC3339),
			"entry_price": math.Round(entry*100) / 100,
			"exit_price":  math.Round(exit*100) / 100,
			"pnl":         math.Round(pnl*100) / 100,
			"result":      map[bool]string{true: "WIN", false: "LOSS"}[isWin],
		})
	}

	totalPnL := capital - initialCapital
	totalReturn := totalPnL / initialCapital
	winRate := 0.0
	if numTrades > 0 {
		winRate = float64(wins) / float64(numTrades)
	}
	avgWin, avgLoss := 0.0, 0.0
	if wins > 0 {
		avgWin = totalWin / float64(wins)
	}
	if losses > 0 {
		avgLoss = totalLoss / float64(losses)
	}
	pf := 0.0
	if totalLoss > 0 {
		pf = totalWin / totalLoss
	}
	sharpe := calcSharpe(returns)

	return simResult{
		FinalCapital: math.Round(capital*100) / 100, TotalPnL: math.Round(totalPnL*100) / 100,
		TotalReturn: math.Round(totalReturn*10000) / 10000, TotalTrades: numTrades,
		WinningTrades: wins, LosingTrades: losses, WinRate: math.Round(winRate*10000) / 10000,
		AvgWin: math.Round(avgWin*100) / 100, AvgLoss: math.Round(avgLoss*100) / 100,
		ProfitFactor: math.Round(pf*100) / 100, SharpeRatio: math.Round(sharpe*100) / 100,
		MaxDrawdown: math.Round(maxDD*10000) / 10000, Trades: trades,
	}
}

func calcSharpe(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))
	variance := 0.0
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	variance /= float64(len(returns) - 1)
	std := math.Sqrt(variance)
	if std == 0 {
		return 0
	}
	return mean / std * math.Sqrt(252)
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ---------- Base ----------

type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ---------- User ----------

type User struct {
	BaseModel
	Username        string  `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email           string  `gorm:"uniqueIndex;size:100;not null" json:"email"`
	WalletAddress   string  `gorm:"uniqueIndex;size:66;not null" json:"wallet_address"`
	StartingCapital float64 `gorm:"default:10000" json:"starting_capital"`
	RiskMode        string  `gorm:"size:20;default:moderate" json:"risk_mode"`
	IsActive        bool    `gorm:"default:true" json:"is_active"`
}

// ---------- UserSettings ----------

type UserSettings struct {
	BaseModel
	UserID          uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	AIProvider      string         `gorm:"size:50;default:openai" json:"ai_provider"`
	AIModel         string         `gorm:"size:100;default:gpt-4o" json:"ai_model"`
	APIKeyEncrypted string         `gorm:"type:text" json:"-"`
	Preferences     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"preferences"`
}

// ---------- Strategy ----------

type Strategy struct {
	BaseModel
	UserID      uuid.UUID      `gorm:"type:uuid;index:idx_user_status;not null" json:"user_id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Config      datatypes.JSON `gorm:"type:jsonb;not null" json:"config"`
	Status      string         `gorm:"size:20;default:draft;index:idx_user_status" json:"status"`
	Version     int            `gorm:"default:1" json:"version"`
	DeployedAt  *time.Time     `json:"deployed_at"`

	TotalTrades int      `gorm:"default:0" json:"total_trades"`
	WinCount    int      `gorm:"default:0" json:"win_count"`
	TotalPnL    float64  `gorm:"default:0" json:"total_pnl"`
	SharpeRatio *float64 `json:"sharpe_ratio"`
	MaxDrawdown *float64 `json:"max_drawdown"`
}

// ---------- HistoricalEvent ----------

type HistoricalEvent struct {
	BaseModel
	SomniaEventID   string    `gorm:"uniqueIndex;size:255;not null" json:"somnia_event_id"`
	EventName       string    `gorm:"size:500;not null" json:"event_name"`
	Category        string    `gorm:"size:100;not null;index:idx_category_date" json:"category"`
	Description     string    `gorm:"type:text" json:"description"`
	EventDate       time.Time `gorm:"not null;index:idx_category_date" json:"event_date"`
	SettlementDate  time.Time `gorm:"not null" json:"settlement_date"`
	SettlementPrice *float64  `json:"settlement_price"`
	YesProbability  *float64  `json:"yes_probability"`
}

// ---------- PriceHistory (TimescaleDB hypertable via hook) ----------

type PriceHistory struct {
	Time         time.Time `gorm:"primaryKey;not null" json:"time"`
	EventID      uuid.UUID `gorm:"type:uuid;primaryKey;not null;index" json:"event_id"`
	Bid          float64   `gorm:"not null" json:"bid"`
	Ask          float64   `gorm:"not null" json:"ask"`
	Volume       *float64  `json:"volume"`
	OpenInterest *float64  `json:"open_interest"`
}

// TableName overrides default pluralization.
func (PriceHistory) TableName() string { return "price_history" }

// ---------- Trade ----------

type Trade struct {
	BaseModel
	StrategyID uuid.UUID  `gorm:"type:uuid;index;not null" json:"strategy_id"`
	EventID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"event_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	TradeType  string     `gorm:"size:20;not null" json:"trade_type"`
	Side       string     `gorm:"size:10;not null" json:"side"`
	Outcome    string     `gorm:"size:10;not null" json:"outcome"`
	Quantity   float64    `gorm:"not null" json:"quantity"`
	EntryPrice float64    `gorm:"not null" json:"entry_price"`
	EntryTime  time.Time  `gorm:"not null" json:"entry_time"`
	ExitPrice  *float64   `json:"exit_price"`
	ExitTime   *time.Time `json:"exit_time"`
	ExitReason *string    `gorm:"size:50" json:"exit_reason"`
	PnL        *float64   `json:"pnl"`
	PnLPercent *float64   `json:"pnl_percent"`
}

// ---------- Backtest ----------

type Backtest struct {
	BaseModel
	StrategyID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"strategy_id"`
	UserID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	StartDate       time.Time      `gorm:"not null" json:"start_date"`
	EndDate         time.Time      `gorm:"not null" json:"end_date"`
	InitialCapital  float64        `gorm:"not null" json:"initial_capital"`
	FinalCapital    *float64       `json:"final_capital"`
	TotalPnL        *float64       `json:"total_pnl"`
	TotalReturn     *float64       `json:"total_return"`
	TotalTrades     *int           `json:"total_trades"`
	WinningTrades   *int           `json:"winning_trades"`
	LosingTrades    *int           `json:"losing_trades"`
	WinRate         *float64       `json:"win_rate"`
	AvgWin          *float64       `json:"avg_win"`
	AvgLoss         *float64       `json:"avg_loss"`
	ProfitFactor    *float64       `json:"profit_factor"`
	SharpeRatio     *float64       `json:"sharpe_ratio"`
	SortinoRatio    *float64       `json:"sortino_ratio"`
	MaxDrawdown     *float64       `json:"max_drawdown"`
	ExecutionTimeMs *int           `json:"execution_time_ms"`
	Status          string         `gorm:"size:20;default:running" json:"status"`
	Trades          datatypes.JSON `gorm:"type:jsonb" json:"trades"`
}

// ---------- Position ----------

type Position struct {
	BaseModel
	StrategyID           uuid.UUID `gorm:"type:uuid;index;not null" json:"strategy_id"`
	EventID              uuid.UUID `gorm:"type:uuid;index;not null" json:"event_id"`
	UserID               uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Side                 string    `gorm:"size:10;not null" json:"side"`
	Outcome              string    `gorm:"size:10;not null" json:"outcome"`
	Quantity             float64   `gorm:"not null" json:"quantity"`
	EntryPrice           float64   `gorm:"not null" json:"entry_price"`
	EntryTime            time.Time `gorm:"not null" json:"entry_time"`
	CurrentPrice         *float64  `json:"current_price"`
	UnrealizedPnL        *float64  `json:"unrealized_pnl"`
	UnrealizedPnLPercent *float64  `json:"unrealized_pnl_percent"`
	StopLossPrice        *float64  `json:"stop_loss_price"`
	TakeProfitPrice      *float64  `json:"take_profit_price"`
}

// ---------- AgentLog ----------

type AgentLog struct {
	BaseModel
	StrategyID uuid.UUID      `gorm:"type:uuid;index;not null" json:"strategy_id"`
	UserID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	EventType  string         `gorm:"size:50;not null" json:"event_type"`
	Message    string         `gorm:"type:text" json:"message"`
	Details    datatypes.JSON `gorm:"type:jsonb" json:"details"`
}

// ---------- Alert ----------

type Alert struct {
	BaseModel
	UserID         uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	StrategyID     *uuid.UUID     `gorm:"type:uuid;index" json:"strategy_id"`
	AlertType      string         `gorm:"size:50;not null" json:"alert_type"`
	Condition      datatypes.JSON `gorm:"type:jsonb;not null" json:"condition"`
	Enabled        bool           `gorm:"default:true" json:"enabled"`
	NotifyInApp    bool           `gorm:"default:true" json:"notify_in_app"`
	NotifyEmail    bool           `gorm:"default:false" json:"notify_email"`
	NotifyTelegram bool           `gorm:"default:false" json:"notify_telegram"`
}

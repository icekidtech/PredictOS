package ai

import (
	"context"
	"encoding/json"
)

// ParseResult is the unified output regardless of provider.
type ParseResult struct {
	Strategy   json.RawMessage `json:"strategy"`
	Confidence float64         `json:"confidence"`
	Warnings   []string        `json:"warnings"`
}

// Provider is the abstraction all AI backends implement.
type Provider interface {
	Name() string
	ParseStrategy(ctx context.Context, naturalLanguage string) (*ParseResult, error)
}

// Registry holds available providers.
type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers ...Provider) *Registry {
	m := make(map[string]Provider)
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &Registry{providers: m}
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.providers))
	for k := range r.providers {
		names = append(names, k)
	}
	return names
}

// StrategyJSONSchema is injected into the system prompt so every provider
// outputs the same shape.
const StrategyJSONSchema = `{
  "triggers": [
    {
      "event_filter": {
        "category": "technology|politics|sports|weather|finance",
        "subcategory": "string (optional)",
        "min_days_to_event": "number (optional)",
        "max_days_to_event": "number (optional)"
      },
      "confidence_threshold": "number 0-1",
      "confidence_source": "internal_model|market_price|ensemble"
    }
  ],
  "actions": {
    "order_type": "limit|market",
    "side": "buy|sell",
    "contract_outcome": "yes|no",
    "position_sizing": {
      "type": "percentage_of_portfolio|fixed|dynamic",
      "value": "number",
      "max_position_size": "number (optional)"
    },
    "entry": { "price": "market or number string" }
  },
  "risk_management": {
    "stop_loss_percent": "number 0-1 (optional)",
    "take_profit_percent": "number 0-1 (optional)",
    "max_drawdown_percent": "number 0-1 (optional)",
    "max_loss_per_day": "number (optional)",
    "position_hold_time_max": "24h|until_event (optional)"
  },
  "execution": {
    "mode": "backtest|live|paused|dry_run",
    "max_concurrent_positions": "number",
    "allowed_times": "24/7|market_hours"
  }
}`

const SystemPrompt = `You are a prediction market strategy parser for PredictOS.
Convert the user's natural language trading strategy into a JSON object that strictly follows this schema:

` + StrategyJSONSchema + `

Rules:
- Output ONLY valid JSON (no markdown, no explanation).
- Include a top-level "_confidence" (0-1) and "_warnings" (string array) alongside the strategy fields.
- If the user omits risk management, set sensible defaults and add a warning.
- Valid categories: technology, politics, sports, weather, finance.
- Confidence threshold defaults to 0.7 if not specified.
`

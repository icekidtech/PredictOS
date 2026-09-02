package dreamdex

import (
	"log"
	"strings"
	"time"

	"predictos-backend/internal/config"
	"predictos-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Syncer polls DreamDEX and upserts into historical_events + price_history.
type Syncer struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewSyncer(db *gorm.DB, cfg *config.Config) *Syncer {
	return &Syncer{db: db, cfg: cfg}
}

// SyncOnce fetches markets for the given network and upserts them.
func (s *Syncer) SyncOnce(network string) (int, error) {
	baseURL := s.cfg.DreamDEXAPIForNetwork(network)
	client := NewClient(baseURL)
	markets, err := client.FetchMarkets()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range markets {
		// Only binary event contracts
		if m.Info.Kind != "" && m.Info.Kind != "binary" {
			continue
		}
		// Use marketId or symbol as somnia_event_id
		eventID := m.MarketID
		if eventID == "" {
			eventID = m.Info.MarketID
		}
		if eventID == "" {
			eventID = m.Symbol
		}
		if eventID == "" {
			continue
		}

		eventName := m.Symbol
		if eventName == "" && len(m.Outcomes) > 0 {
			eventName = m.Outcomes[0].Symbol
		}
		if eventName == "" {
			eventName = eventID
		}

		category := inferCategory(m)
		expiry := parseExpiry(m.Info.Expiry)

		ev := models.HistoricalEvent{
			BaseModel:      models.BaseModel{ID: uuid.New()},
			SomniaEventID:  eventID,
			Network:        network,
			EventName:      eventName,
			Category:       category,
			Description:    describeMarket(m),
			EventDate:      expiry,
			SettlementDate: expiry,
		}

		// Upsert by (somnia_event_id, network)
		var existing models.HistoricalEvent
		err := s.db.Where("somnia_event_id = ? AND network = ?", eventID, network).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := s.db.Create(&ev).Error; err != nil {
				log.Printf("dreamdex sync: create %s: %v", eventID, err)
				continue
			}
			count++
		} else if err == nil {
			// Update mutable fields
			s.db.Model(&existing).Updates(map[string]interface{}{
				"event_name": eventName,
				"category":   category,
			})
			ev.ID = existing.ID
		}

		// Price snapshot — try embedded book first, then fetch per-symbol
		var bid, ask *float64
		if m.OrderBook != nil && len(m.OrderBook.Bids) > 0 && len(m.OrderBook.Asks) > 0 {
			b := m.OrderBook.Bids[0][0]
			a := m.OrderBook.Asks[0][0]
			bid, ask = &b, &a
		} else {
			// Fetch order book for this market's outcome symbol
			symbol := m.Symbol
			if symbol == "" && len(m.Outcomes) > 0 {
				symbol = m.Outcomes[0].Symbol
			}
			if symbol != "" {
				if bids, asks, err := client.FetchOrderBook(symbol); err == nil && len(bids) > 0 && len(asks) > 0 {
					b := bids[0][0]
					a := asks[0][0]
					bid, ask = &b, &a
				}
			}
		}
		if bid != nil && ask != nil {
			ph := models.PriceHistory{
				Time:    time.Now().UTC(),
				EventID: ev.ID,
				Bid:     *bid,
				Ask:     *ask,
			}
			if existing.ID != uuid.Nil {
				ph.EventID = existing.ID
			}
			_ = s.db.Create(&ph).Error
		}
	}

	return count, nil
}

// StartPolling runs SyncOnce every interval for the given network.
func (s *Syncer) StartPolling(network string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			n, err := s.SyncOnce(network)
			if err != nil {
				log.Printf("dreamdex sync (%s): %v", network, err)
			} else if n > 0 {
				log.Printf("dreamdex sync (%s): upserted %d markets", network, n)
			}
			<-ticker.C
		}
	}()
}

func inferCategory(m Market) string {
	sym := strings.ToLower(m.Symbol + " " + m.BaseSymbol + " " + m.Info.Underlying)
	switch {
	case strings.Contains(sym, "btc") || strings.Contains(sym, "eth") || strings.Contains(sym, "sol"):
		return "crypto"
	case strings.Contains(sym, "up") || strings.Contains(sym, "down"):
		return "crypto"
	default:
		return "crypto"
	}
}

func describeMarket(m Market) string {
	if m.Info.Underlying != "" && m.Info.Strike != "" {
		return m.Info.Underlying + " " + m.Info.Strike + " — binary Up/Down"
	}
	return "DreamDEX event contract — " + m.Symbol
}

func parseExpiry(s string) time.Time {
	if s == "" {
		return time.Now().Add(24 * time.Hour)
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z", s); err == nil {
		return t
	}
	return time.Now().Add(24 * time.Hour)
}

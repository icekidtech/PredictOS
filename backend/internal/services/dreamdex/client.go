package dreamdex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client talks to DreamDEX REST API (stg.api.dreamdex.io / api.dreamdex.io).
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Market mirrors the DreamDEX /v0/markets response.
type Market struct {
	MarketID    string `json:"marketId"`
	Symbol      string `json:"symbol"`
	BaseSymbol  string `json:"baseSymbol"`
	QuoteSymbol string `json:"quoteSymbol"`
	Status      int    `json:"status"` // 1 = Trading
	Active      bool   `json:"active"`
	Info        struct {
		MarketID   string `json:"marketId"`
		Kind       string `json:"kind"` // "binary"
		Expiry     string `json:"expiry"`
		Strike     string `json:"strike"`
		Underlying string `json:"underlying"`
	} `json:"info"`
	Outcomes []struct {
		Symbol string `json:"symbol"`
	} `json:"outcomes"`
	OrderBook *struct {
		Bids [][]float64 `json:"bids"`
		Asks [][]float64 `json:"asks"`
	} `json:"orderBook"`
}

type marketsResponse struct {
	Markets map[string]Market `json:"markets"`
	// Some versions return array
	MarketsArray []Market `json:"-"`
}

func (c *Client) FetchMarkets() ([]Market, error) {
	resp, err := c.http.Get(c.baseURL + "/markets")
	if err != nil {
		return nil, fmt.Errorf("dreamdex markets: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("dreamdex markets %d: %s", resp.StatusCode, string(body))
	}

	// Try map form first
	var mResp marketsResponse
	if err := json.Unmarshal(body, &mResp); err == nil && len(mResp.Markets) > 0 {
		out := make([]Market, 0, len(mResp.Markets))
		for _, m := range mResp.Markets {
			out = append(out, m)
		}
		return out, nil
	}
	// Try array form
	var arr []Market
	if err := json.Unmarshal(body, &arr); err == nil && len(arr) > 0 {
		return arr, nil
	}
	// Try wrapped { data: [...] } or { markets: [...] }
	var wrapped struct {
		Data    []Market `json:"data"`
		Markets []Market `json:"markets"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		if len(wrapped.Data) > 0 {
			return wrapped.Data, nil
		}
		if len(wrapped.Markets) > 0 {
			return wrapped.Markets, nil
		}
	}
	return nil, fmt.Errorf("dreamdex: unexpected markets response: %s", string(body[:min(500, len(body))]))
}

func (c *Client) FetchOrderBook(symbol string) (bids [][]float64, asks [][]float64, err error) {
	// DreamDEX order book endpoint — try common paths
	paths := []string{
		"/orderbook?symbol=" + symbol,
		"/orderBook?symbol=" + symbol,
		"/book?symbol=" + symbol,
	}
	for _, p := range paths {
		resp, e := c.http.Get(c.baseURL + p)
		if e != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			continue
		}
		var ob struct {
			Bids [][]float64 `json:"bids"`
			Asks [][]float64 `json:"asks"`
		}
		if json.Unmarshal(body, &ob) == nil && (len(ob.Bids) > 0 || len(ob.Asks) > 0) {
			return ob.Bids, ob.Asks, nil
		}
	}
	return nil, nil, fmt.Errorf("orderbook not available for %s", symbol)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

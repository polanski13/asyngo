package oneof

import "time"

type TickerPayload struct {
	EventType string    `json:"eventType" enum:"ticker"`
	Symbol    string    `json:"symbol" example:"BTC-USD" validate:"required"`
	Price     float64   `json:"price" example:"43250.50" minimum:"0"`
	Timestamp time.Time `json:"timestamp"`
}

type OrderBookPayload struct {
	EventType string `json:"eventType" enum:"orderBook"`
	Symbol    string `json:"symbol" validate:"required"`
	Depth     int    `json:"depth" example:"10"`
}

type TradePayload struct {
	EventType string  `json:"eventType" enum:"trade"`
	Symbol    string  `json:"symbol" validate:"required"`
	Price     float64 `json:"price" minimum:"0"`
	Quantity  float64 `json:"quantity" minimum:"0"`
}

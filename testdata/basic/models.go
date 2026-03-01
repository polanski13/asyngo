package basic

import "time"

type TickerPayload struct {
	Symbol    string    `json:"symbol" example:"BTC-USD" validate:"required"`
	Price     float64   `json:"price" example:"43250.50" minimum:"0"`
	Volume24h float64   `json:"volume24h" example:"1234567.89" minimum:"0"`
	Change24h float64   `json:"change24h" example:"-2.35"`
	Timestamp time.Time `json:"timestamp"`
	Side      string    `json:"side" enum:"buy,sell"`
}

type OrderBookPayload struct {
	Symbol string       `json:"symbol" validate:"required"`
	Bids   []PriceLevel `json:"bids"`
	Asks   []PriceLevel `json:"asks"`
	SeqNum int64        `json:"seqNum"`
}

type PriceLevel struct {
	Price    float64 `json:"price" example:"43250.50" minimum:"0"`
	Quantity float64 `json:"quantity" example:"1.5" minimum:"0"`
}

type SubscribeRequest struct {
	Action  string   `json:"action" enum:"subscribe" validate:"required"`
	Streams []string `json:"streams"`
}

type SubscriptionAck struct {
	Status            string   `json:"status" enum:"ok,error"`
	SubscribedStreams []string `json:"subscribedStreams"`
	Error             string   `json:"error,omitempty"`
}

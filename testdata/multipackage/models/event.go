package models

import "time"

type Event struct {
	ID        string    `json:"id" validate:"required"`
	Type      string    `json:"type" enum:"created,updated,deleted"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type Notification struct {
	Event
	UserID  string `json:"userId" validate:"required"`
	Channel string `json:"channel"`
	Read    bool   `json:"read"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message" validate:"required"`
}

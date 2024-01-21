package api

import (
	"time"
)

type commonFields struct {
	Mac       string    `json:"mac"`
	Timestamp time.Time `json:"timestamp"`
}

type CardEvent struct {
	CardID string `json:"card_id"`
	commonFields
}

type MoveEvent struct {
	commonFields
}

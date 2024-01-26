package api

type commonFields struct {
	Mac       string `json:"mac"`
	Timestamp int64  `json:"timestamp"`
}

type CardEvent struct {
	CardID string `json:"card_id"`
	commonFields
}

type MoveEvent struct {
	commonFields
}

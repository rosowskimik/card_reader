package api

type commonFields struct {
	Id string `json:"systemId"`
}

type CardEvent struct {
	commonFields
	CardID string `json:"cardValue"`
}

type MoveEvent struct {
	commonFields
	Timestamp int64 `json:"timestamp"`
}

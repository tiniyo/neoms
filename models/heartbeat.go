package models

type HeartBeatPrivRequest struct {
	AccountID string  `json:"accountid"`
	CallID    string  `json:"callid"`
	Rate      float64 `json:"rate"`
	Pulse     int64   `json:"pulse"`
	Duration  int64   `json:"duration"`
}

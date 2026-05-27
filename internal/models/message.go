package models

type Message struct {
	Action   string `json:"action"`
	Text     string `json:"text"`
	TargetID int64  `json:"target_id"`
}

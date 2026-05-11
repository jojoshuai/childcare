package model

import "time"

type Measurement struct {
	ID         string    `json:"id"`
	ChildID    string    `json:"child_id"`
	Type       string    `json:"type"` // "weight", "height"
	Value      float64   `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	Note       *string   `json:"note"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

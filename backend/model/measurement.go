package model

import "time"

type Measurement struct {
	ID         string
	ChildID    string
	Type       string  // "weight", "height", "head_circumference"
	Value      float64
	MeasuredAt time.Time
	Note       *string
	CreatedBy  string
	CreatedAt  time.Time
}

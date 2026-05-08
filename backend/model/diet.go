package model

import "time"

type DietRecord struct {
	ID          string     `json:"id"`
	ChildID     string     `json:"child_id"`
	FoodName    string     `json:"food_name"`
	FoodType    string     `json:"food_type"`
	AmountLevel int        `json:"amount_level"`
	RecordTime  time.Time  `json:"record_time"`
	Notes       *string    `json:"notes"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

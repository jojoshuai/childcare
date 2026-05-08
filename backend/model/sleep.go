package model

import "time"

type SleepRecord struct {
	ID        string     `json:"id"`
	ChildID   string     `json:"child_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	WokeUp    bool       `json:"woke_up"`
	WakeCount int        `json:"wake_count"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
}

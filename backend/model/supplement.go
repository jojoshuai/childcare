package model

import "time"

type SupplementRecord struct {
	ID             string     `json:"id"`
	ChildID        string     `json:"child_id"`
	SupplementName string     `json:"supplement_name"`
	Dose           *string    `json:"dose"`
	TakenAt        time.Time  `json:"taken_at"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
}

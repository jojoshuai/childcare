package model

import "time"

type Child struct {
	ID        string    `json:"id"`
	FamilyID  string    `json:"family_id"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"` // "male" or "female"
	BirthDate time.Time `json:"birth_date"`
	CreatedAt time.Time `json:"created_at"`
}

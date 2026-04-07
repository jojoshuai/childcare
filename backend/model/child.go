package model

import "time"

type Child struct {
	ID        string
	FamilyID  string
	Name      string
	Gender    string    // "male" or "female"
	BirthDate time.Time
	CreatedAt time.Time
}

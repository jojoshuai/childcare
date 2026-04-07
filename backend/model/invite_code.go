package model

import "time"

type InviteCode struct {
	ID        string
	FamilyID  string
	Code      string
	ExpiresAt time.Time
	Used      bool
	CreatedBy string
	CreatedAt time.Time
}

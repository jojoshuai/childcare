package model

import "time"

type InviteCode struct {
	ID        string    `json:"id"`
	FamilyID  string    `json:"family_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

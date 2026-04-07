package model

import "time"

type User struct {
	ID           string    `json:"id"`
	FamilyID     *string   `json:"family_id"`  // NULL when miniprogram user hasn't joined a family
	Username     *string   `json:"username"`   // NULL for wx-only users
	PasswordHash *string   `json:"-"`          // never serialized to client
	WxOpenID     *string   `json:"-"`          // never serialized to client
	Nickname     string    `json:"nickname"`
	Role         *string   `json:"role"`       // NULL when no family; "owner" or "member"
	CreatedAt    time.Time `json:"created_at"`
}

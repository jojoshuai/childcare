package model

import "time"

type User struct {
	ID           string
	FamilyID     *string   // NULL when miniprogram user hasn't joined a family
	Username     *string   // NULL for wx-only users
	PasswordHash *string   // NULL for wx-only users
	WxOpenID     *string   // NULL for web-only users
	Nickname     string
	Role         *string   // NULL when no family; "owner" or "member"
	CreatedAt    time.Time
}

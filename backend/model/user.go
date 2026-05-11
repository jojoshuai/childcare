package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     *string   `json:"username"`
	PasswordHash *string   `json:"-"`
	WxOpenID     *string   `json:"-"`
	Nickname     string    `json:"nickname"`
	CreatedAt    time.Time `json:"created_at"`
}

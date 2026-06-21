package channel

import "time"

type Channel struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

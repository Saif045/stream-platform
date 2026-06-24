package channel

import "time"

type PublicChannel struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Channel struct {
	PublicChannel

	UserID string `json:"user_id"`
}

func (c *Channel) Public() any {
	if c == nil {
		return nil
	}

	return c.PublicChannel
}

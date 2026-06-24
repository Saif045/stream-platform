package user

import "time"

type PublicUser struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	PublicUser

	PasswordHash string `json:"-"`
}

func (u *User) Public() any {
	if u == nil {
		return nil
	}

	return u.PublicUser
}

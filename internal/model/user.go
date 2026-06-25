package model

import "time"

// User is a form creator, identified by their Google account.
type User struct {
	ID         string    `json:"id"`
	GoogleSub  string    `json:"-"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	PictureURL string    `json:"picture_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GoogleProfile is the subset of the Google userinfo response we persist.
type GoogleProfile struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

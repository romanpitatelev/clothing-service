package entity

import "time"

type Brand struct {
	ID        BrandID   `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Logo      string    `json:"logo" db:"logo"`
	URL       string    `json:"url" db:"url"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type RequestedBrand struct {
	Brand   string `json:"brand" db:"brand"`
	Comment string `json:"comment" db:"comment"`
	UserID  UserID `json:"userId" db:"user_id"`
}

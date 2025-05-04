package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrClothingNotFound = errors.New("clothing not found")

type ClothingID uuid.UUID

func (c ClothingID) String() string {
	return uuid.UUID(c).String()
}

func (c *ClothingID) UnmarshalText(data []byte) error {
	return (*uuid.UUID)(c).UnmarshalText(data)
}

func (c ClothingID) MarshalText() ([]byte, error) {
	return uuid.UUID(c).MarshalText()
}

type RelatedProduct struct {
	URL   *string `json:"url"`
	Photo *string `json:"photo"`
}

type Clothing struct {
	ID              ClothingID       `json:"id" db:"id"`
	Brand           string           `json:"brand" db:"brand"`
	Category        string           `json:"category" db:"category"`
	Color           string           `json:"color" db:"color"`
	Name            string           `json:"name" db:"name"`
	Price           float64          `json:"price" db:"price"`
	Sizes           string           `json:"sizes" db:"sizes"`
	ProductURL      string           `json:"productUrl" db:"product_url"`
	PrimaryPhoto    string           `json:"primaryPhoto" db:"primary_photo"`
	ExtraPhotos     []*string        `json:"extraPhotos" db:"extra_photos"`
	RelatedProducts []RelatedProduct `json:"relatedProducts" db:"related_products"`
	CreatedAt       time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time        `json:"updatedAt" db:"updated_at"`
}

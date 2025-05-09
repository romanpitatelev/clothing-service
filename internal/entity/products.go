package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrVariantNotFound = errors.New("variant not found")

type VariantID uuid.UUID

func (v VariantID) String() string {
	return uuid.UUID(v).String()
}

func (v *VariantID) UnmarshalText(data []byte) error {
	return (*uuid.UUID)(v).UnmarshalText(data)
}

func (v VariantID) MarshalText() ([]byte, error) {
	return uuid.UUID(v).MarshalText()
}

type ProductID uuid.UUID

func (p ProductID) String() string {
	return uuid.UUID(p).String()
}

func (p *ProductID) UnmarshalText(data []byte) error {
	return (*uuid.UUID)(p).UnmarshalText(data)
}

func (p ProductID) MarshalText() ([]byte, error) {
	return uuid.UUID(p).MarshalText()
}

type BrandID uuid.UUID

func (b BrandID) String() string {
	return uuid.UUID(b).String()
}

func (b *BrandID) UnmarshalText(data []byte) error {
	return (*uuid.UUID)(b).UnmarshalText(data)
}

func (b BrandID) MarshalText() ([]byte, error) {
	return uuid.UUID(b).MarshalText()
}

type RelatedProduct struct {
	URL   *string `json:"url"`
	Photo *string `json:"photo"`
}

type Variant struct {
	ID                      VariantID        `json:"id" db:"id"`
	ProductID               ProductID        `json:"productId" db:"product_id"`
	BrandID                 BrandID          `json:"brandId" db:"brand_id"`
	Brand                   string           `json:"brand" db:"brand"`
	Category                string           `json:"category" db:"category"`
	Color                   string           `json:"color" db:"color"`
	ColorHex                string           `json:"colorHex" db:"color_hex"`
	Name                    string           `json:"name" db:"name"`
	Price                   int              `json:"price" db:"price"`
	Currency                string           `json:"currency" db:"currency"`
	Sizes                   []string         `json:"sizes" db:"sizes"`
	SoldOutSizes            []string         `json:"soldOutSizes" db:"sold_out_sizes"`
	ProductURL              string           `json:"productUrl" db:"product_url"`
	PrimaryPhoto            string           `json:"primaryPhoto" db:"primary_photo"`
	PrimaryPhotoOriginalURL string           `json:"-" db:"-"`
	Photos                  []string         `json:"photos" db:"photos"`
	PhotoOriginalURL        []string         `json:"-" db:"-"`
	RelatedProducts         []RelatedProduct `json:"relatedProducts" db:"related_products"`
	CreatedAt               time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt               time.Time        `json:"updatedAt" db:"updated_at"`
}

type Product struct {
	ID        ProductID `json:"id" db:"id"`
	BrandID   BrandID   `json:"brandId" db:"brand_id"`
	Category  string    `json:"category" db:"category"`
	Name      string    `json:"name" db:"name"`
	Price     int       `json:"price" db:"price"`
	Currency  string    `json:"currency" db:"currency"`
	Colors    []Color   `json:"colors" db:"colors"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type Color struct {
	Name string `json:"name" db:"name"`
	Hex  string `json:"hex" db:"hex"`
}

type ListRequest struct {
	Sorting    string
	Descending bool
	Limit      int
	Offset     int
	Text       string
}

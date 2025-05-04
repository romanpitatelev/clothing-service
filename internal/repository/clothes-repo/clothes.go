package clothesrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
)

type database interface {
	Exec(ctx context.Context, sq string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sq string, arguments ...any) (pgx.Rows, error)
	GetTXFromContext(ctx context.Context) store.Transaction
}

type Repo struct {
	db database
}

func New(db database) *Repo {
	return &Repo{
		db: db,
	}
}

const selectClothingQuery = `
SELECT id,
       brand,
       category,
       color,
       name,
       price,
       sizes,
       product_url,
       primary_photo,
       array_remove(ARRAY [extra_photo_1,extra_photo_2,extra_photo_3,extra_photo_4,extra_photo_5],
                    null)                                                                     AS extra_photos,
       ARRAY [json_build_object('url', related_product_url_1, 'photo', related_product_photo_1),
           json_build_object('url', related_product_url_2, 'photo', related_product_photo_2),
           json_build_object('url', related_product_url_3, 'photo', related_product_photo_3),
           json_build_object('url', related_product_url_4, 'photo', related_product_photo_4),
           json_build_object('url', related_product_url_5, 'photo', related_product_photo_5)] AS related_products,
       created_at,
       updated_at
FROM clothes
WHERE TRUE
  AND deleted_at IS NULL
`

func (r *Repo) GetClothing(ctx context.Context, id entity.ClothingID) (entity.Clothing, error) {
	db := r.db.GetTXFromContext(ctx)

	var clothing entity.Clothing

	if err := pgxscan.Get(ctx, db, &clothing, selectClothingQuery+"AND id = $1", id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Clothing{}, entity.ErrClothingNotFound
		}

		return entity.Clothing{}, fmt.Errorf("error getting clothing: %w", err)
	}

	for i := range clothing.RelatedProducts {
		if clothing.RelatedProducts[i].Photo == nil && clothing.RelatedProducts[i].URL == nil {
			clothing.RelatedProducts = clothing.RelatedProducts[:i]

			break
		}
	}

	return clothing, nil
}

const (
	extraElemsMaxLen   = 5
	clothingStructSize = 24
)

const insertClothingQuery = `
INSERT INTO clothes (id, brand, category, color, name, price, sizes, product_url, primary_photo, extra_photo_1, extra_photo_2,
                     extra_photo_3, extra_photo_4, extra_photo_5, related_product_url_1,
                     related_product_url_2, related_product_url_3, related_product_url_4, related_product_url_5,
                     related_product_photo_1, related_product_photo_2, related_product_photo_3,
                     related_product_photo_4, related_product_photo_5)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
ON CONFLICT (product_url) WHERE deleted_at IS NULL DO UPDATE
SET brand = excluded.brand,
    category = excluded.category,
    color = excluded.color,
    name = excluded.name,
    price = excluded.price,
    sizes = excluded.sizes,
    product_url = excluded.product_url,
    primary_photo = excluded.primary_photo,
    extra_photo_1 = excluded.extra_photo_1,
    extra_photo_2 = excluded.extra_photo_2,
    extra_photo_3 = excluded.extra_photo_3,
    extra_photo_4 = excluded.extra_photo_4,
    extra_photo_5 = excluded.extra_photo_5,
    related_product_url_1 = excluded.related_product_url_1,
    related_product_url_2 = excluded.related_product_url_2,
    related_product_url_3 = excluded.related_product_url_3,
    related_product_url_4 = excluded.related_product_url_4,
    related_product_url_5 = excluded.related_product_url_5,
    related_product_photo_1 = excluded.related_product_photo_1,
    related_product_photo_2 = excluded.related_product_photo_2,
    related_product_photo_3 = excluded.related_product_photo_3,
    related_product_photo_4 = excluded.related_product_photo_4,
    related_product_photo_5 = excluded.related_product_photo_5
`

func (r *Repo) UpsertClothing(ctx context.Context, clothing entity.Clothing) error {
	db := r.db.GetTXFromContext(ctx)

	args := make([]any, 0, clothingStructSize)

	args = append(args, clothing.ID, clothing.Brand, clothing.Category, clothing.Color, clothing.Name, clothing.Price,
		clothing.Sizes, clothing.ProductURL, clothing.PrimaryPhoto)

	for i := range extraElemsMaxLen {
		if i < len(clothing.ExtraPhotos) {
			args = append(args, clothing.ExtraPhotos[i])
		} else {
			args = append(args, nil)
		}
	}

	for i := range extraElemsMaxLen {
		if i < len(clothing.RelatedProducts) {
			args = append(args, clothing.RelatedProducts[i].URL)
		} else {
			args = append(args, nil)
		}
	}

	for i := range extraElemsMaxLen {
		if i < len(clothing.RelatedProducts) {
			args = append(args, clothing.RelatedProducts[i].Photo)
		} else {
			args = append(args, nil)
		}
	}

	if _, err := db.Exec(ctx, insertClothingQuery, args...); err != nil {
		return fmt.Errorf("error upserting clothing: %w", err)
	}

	return nil
}

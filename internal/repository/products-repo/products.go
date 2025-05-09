package clothesrepo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

const selectVariantQuery = `
SELECT id,
       brand,
       category,
       color,
       name,
       price,
       sizes,
       product_url,
       primary_photo,
       array_remove(ARRAY [photo_1, photo_2, photo_3, photo_4, photo_5, photo_6, photo_7, photo_8, photo_9, photo_10],
                    null)                                                                     AS photos,
       ARRAY [json_build_object('url', related_product_url_1, 'photo', related_product_photo_1),
           json_build_object('url', related_product_url_2, 'photo', related_product_photo_2),
           json_build_object('url', related_product_url_3, 'photo', related_product_photo_3),
           json_build_object('url', related_product_url_4, 'photo', related_product_photo_4),
           json_build_object('url', related_product_url_5, 'photo', related_product_photo_5)] AS related_products,
       created_at,
       updated_at
FROM variants
WHERE TRUE
  AND deleted_at IS NULL
`

func (r *Repo) GetVariant(ctx context.Context, id entity.VariantID) (entity.Variant, error) {
	db := r.db.GetTXFromContext(ctx)

	var variant entity.Variant

	if err := pgxscan.Get(ctx, db, &variant, selectVariantQuery+"AND id = $1", id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Variant{}, entity.ErrVariantNotFound
		}

		return entity.Variant{}, fmt.Errorf("error getting variant: %w", err)
	}

	for i := range variant.RelatedProducts {
		if variant.RelatedProducts[i].Photo == nil && variant.RelatedProducts[i].URL == nil {
			variant.RelatedProducts = variant.RelatedProducts[:i]

			break
		}
	}

	return variant, nil
}

const (
	photosMaxLen          = 10
	relatedProductsMaxLen = 5
	variantStructSize     = 33
)

const upsertVariantQuery = `
INSERT INTO variants (id, product_id, color, color_hex, sizes, sold_out_sizes,
                     product_url, primary_photo, photo_1, photo_2, photo_3, photo_4, photo_5, photo_6, photo_7, photo_8, 
                     photo_9, photo_10, related_product_url_1, related_product_url_2, related_product_url_3,
                     related_product_url_4, related_product_url_5, related_product_photo_1, related_product_photo_2, 
                     related_product_photo_3, related_product_photo_4, related_product_photo_5)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, 
        $24, $25, $26, $27, $28)
ON CONFLICT (product_url) WHERE deleted_at IS NULL DO UPDATE
SET product_id = excluded.product_id,
    color = excluded.color,
    color_hex = excluded.color_hex,
    sizes = excluded.sizes,
    product_url = excluded.product_url,
    primary_photo = excluded.primary_photo,
    photo_1 = excluded.photo_1,
    photo_2 = excluded.photo_2,
    photo_3 = excluded.photo_3,
    photo_4 = excluded.photo_4,
    photo_5 = excluded.photo_5,
    photo_6 = excluded.photo_6,
    photo_7 = excluded.photo_7,
    photo_8 = excluded.photo_8,
    photo_9 = excluded.photo_9,
    photo_10 = excluded.photo_10,
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

func (r *Repo) UpsertVariant(ctx context.Context, variant entity.Variant) error {
	db := r.db.GetTXFromContext(ctx)

	args := make([]any, 0, variantStructSize)

	args = append(args, variant.ID, variant.ProductID, variant.Color, variant.ColorHex, variant.Sizes,
		variant.SoldOutSizes, variant.ProductURL, variant.PrimaryPhoto)

	for i := range photosMaxLen {
		if i < len(variant.Photos) {
			args = append(args, variant.Photos[i])
		} else {
			args = append(args, nil)
		}
	}

	for i := range relatedProductsMaxLen {
		if i < len(variant.RelatedProducts) {
			args = append(args, variant.RelatedProducts[i].URL)
		} else {
			args = append(args, nil)
		}
	}

	for i := range relatedProductsMaxLen {
		if i < len(variant.RelatedProducts) {
			args = append(args, variant.RelatedProducts[i].Photo)
		} else {
			args = append(args, nil)
		}
	}

	if _, err := db.Exec(ctx, upsertVariantQuery, args...); err != nil {
		return fmt.Errorf("error upserting variant: %w", err)
	}

	return nil
}

const upsertProductQuery = `
INSERT INTO products (id, brand_id, category, name, price, currency, colors)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (brand_id, category, name) WHERE deleted_at IS NULL
    DO UPDATE 
    SET price = excluded.price,
    	currency = excluded.currency,
    	colors = excluded.colors
RETURNING id, brand_id, category, name, price, currency, colors, created_at, updated_at
`

func (r *Repo) UpsertProduct(ctx context.Context, product entity.Product) (entity.Product, error) {
	db := r.db.GetTXFromContext(ctx)

	if err := pgxscan.Get(ctx, db, &product, upsertProductQuery, product.ID, product.BrandID, product.Category,
		product.Name, product.Price, product.Currency, product.Colors); err != nil {
		return entity.Product{}, fmt.Errorf("error upserting product: %w", err)
	}

	return product, nil
}

const upsertBrandQuery = `
INSERT INTO brands (id, name)
VALUES ($1, $2)
ON CONFLICT (name) WHERE deleted_at IS NULL DO UPDATE SET updated_at = NOW()
RETURNING id, name, logo, url, created_at, updated_at`

func (r *Repo) UpsertBrand(ctx context.Context, brand entity.Brand) (entity.Brand, error) {
	db := r.db.GetTXFromContext(ctx)

	if err := pgxscan.Get(ctx, db, &brand, upsertBrandQuery, brand.ID, brand.Name); err != nil {
		return entity.Brand{}, fmt.Errorf("error upserting brand: %w", err)
	}

	return brand, nil
}

const listBrandsQuery = `
SELECT id, name, logo, url, created_at, updated_at
FROM brands
WHERE deleted_at IS NULL
`

const defaultLimit = 25

func (r *Repo) ListBrands(ctx context.Context, req entity.ListRequest) ([]entity.Brand, error) {
	db := r.db.GetTXFromContext(ctx)

	mapping := map[string]string{
		"name":      "name",
		"url":       "url",
		"createdAt": "created_at",
	}

	var args []any

	builder := strings.Builder{}
	builder.WriteString(listBrandsQuery)

	if req.Text == "" {
		args = append(args, "%"+req.Text+"%")
		builder.WriteString(fmt.Sprintf(`  AND concat_ws('', name, url) ILIKE $%d`, len(args)))
	}

	orderBy := mapping[req.Sorting]
	if orderBy == "" {
		orderBy = mapping["name"]
	}

	builder.WriteString(" ORDER BY " + orderBy)

	if req.Descending {
		builder.WriteString(" DESC")
	}

	limit := defaultLimit
	if req.Limit > 0 {
		limit = req.Limit
	}

	builder.WriteString(" LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(req.Offset))

	var result []entity.Brand

	if err := pgxscan.Select(ctx, db, &result, builder.String(), args...); err != nil {
		return nil, fmt.Errorf("error listing brands: %w", err)
	}

	return result, nil
}

const setRelatedProductPhotoTmpl = `
UPDATE variants
SET related_product_photo_%[1]d = v.primary_photo
FROM (SELECT product_url, primary_photo FROM variants) v
WHERE related_product_url_%[1]d = v.product_url
`

func (r *Repo) SetRelatedPhotos(ctx context.Context) error {
	for i := range relatedProductsMaxLen {
		if _, err := r.db.Exec(ctx, fmt.Sprintf(setRelatedProductPhotoTmpl, i+1)); err != nil {
			return fmt.Errorf("error setting related_product_photo: %w", err)
		}
	}

	return nil
}

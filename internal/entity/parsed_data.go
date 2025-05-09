package entity

import (
	"strconv"
	"strings"
	"time"

	"github.com/essentialkaos/translit/v3"
	"github.com/google/uuid"
)

type JSONImageMapping map[string]Mapping

type Mapping struct {
	Path string    `json:"path"`
	Date time.Time `json:"date"`
}

type JSONProduct struct {
	Variants []JSONVariant `json:"variants"`
	GrabDate time.Time     `json:"grabDate"`
}

type JSONVariant struct {
	BestPhotoUrls []string `json:"bestPhotoUrls"`
	BrandName     string   `json:"brandName"`
	CategoryName  string   `json:"categoryName"`
	ClothingName  string   `json:"clothingName"`
	URL           string   `json:"url"`
	Color         struct {
		Hex  string `json:"hex"`
		Name string `json:"name"`
	} `json:"color"`
	Price struct {
		Amount   int    `json:"amount"`
		Currency string `json:"currency"`
	} `json:"price"`
	Sizes               []string `json:"sizes"`
	SoldOutSizes        []string `json:"soldOutSizes"`
	RelatedClothingUrls []string `json:"relatedClothingUrls"`
	PhotoUrls           []string `json:"photoUrls"`
}

func (c JSONProduct) Products() (Product, []Variant) {
	product := Product{
		ID:     ProductID(uuid.New()),
		Colors: make([]Color, len(c.Variants)),
	}
	variants := make([]Variant, len(c.Variants))

	colors := map[string]bool{}

	for i, variant := range c.Variants {
		if !colors[variant.Color.Name] {
			product.Colors[i] = Color{
				Name: variant.Color.Name,
				Hex:  variant.Color.Hex,
			}

			colors[variant.Color.Name] = true
		}

		variants[i].ID = VariantID(uuid.New())
		variants[i].ProductID = product.ID
		variants[i].Brand = variant.BrandName
		variants[i].Category = variant.CategoryName
		variants[i].Color = variant.Color.Name
		variants[i].ColorHex = variant.Color.Hex
		variants[i].Name = variant.ClothingName
		variants[i].Price = variant.Price.Amount
		variants[i].Currency = variant.Price.Currency
		variants[i].Sizes = variant.Sizes
		variants[i].SoldOutSizes = variant.SoldOutSizes
		variants[i].ProductURL = variant.URL

		variants[i].PhotoOriginalURL = make([]string, len(variant.PhotoUrls))
		variants[i].Photos = make([]string, len(variant.PhotoUrls))

		for j := range variant.PhotoUrls {
			variants[i].PhotoOriginalURL[j] = variant.PhotoUrls[j]
			variants[i].Photos[j] = variant.photoName(j)
		}

		if len(variant.BestPhotoUrls) > 0 {
			variants[i].PrimaryPhotoOriginalURL = variant.BestPhotoUrls[0]
			variants[i].PrimaryPhoto = variant.photoName(getIndex(variant.PhotoUrls, variant.BestPhotoUrls[0]))
		}

		variants[i].RelatedProducts = make([]RelatedProduct, len(variant.RelatedClothingUrls))
		for j := range variant.RelatedClothingUrls {
			variants[i].RelatedProducts[j] = RelatedProduct{
				URL: &variant.RelatedClothingUrls[j],
			}
		}
	}

	if len(variants) > 0 {
		product.Name = variants[0].Name
		product.Category = variants[0].Category
		product.Price = variants[0].Price
		product.Currency = variants[0].Currency
	}

	return product, variants
}

func (v JSONVariant) photoName(i int) string {
	parts := strings.Split(v.PhotoUrls[i], ".")

	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(translit.BGN(
		v.BrandName+"/"+
			v.CategoryName+"/"+
			v.Color.Name+"/"+
			v.ClothingName)+"/"+
		strconv.Itoa(i), " ", "_"), "′", ""), "%", "") + "." +
		parts[len(parts)-1]
}

func getIndex(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}

	return -1
}

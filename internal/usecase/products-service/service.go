package clothesservice

import (
	"context"
	"fmt"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type productsStore interface {
	GetVariant(ctx context.Context, id entity.VariantID) (entity.Variant, error)
}

type Service struct {
	productsStore productsStore
}

func New(productsStore productsStore) *Service {
	return &Service{
		productsStore: productsStore,
	}
}

func (s *Service) GetVariant(ctx context.Context, id entity.VariantID) (entity.Variant, error) {
	variant, err := s.productsStore.GetVariant(ctx, id)
	if err != nil {
		return entity.Variant{}, fmt.Errorf("GetVariant %w", err)
	}

	return variant, nil
}

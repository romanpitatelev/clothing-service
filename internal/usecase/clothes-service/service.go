package clothesservice

import (
	"context"
	"fmt"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type clothesStore interface {
	GetClothing(ctx context.Context, id entity.ClothingID) (entity.Clothing, error)
}

type Service struct {
	clothesStore clothesStore
}

func New(clothesStore clothesStore) *Service {
	return &Service{
		clothesStore: clothesStore,
	}
}

func (s *Service) GetClothing(ctx context.Context, id entity.ClothingID) (entity.Clothing, error) {
	clothing, err := s.clothesStore.GetClothing(ctx, id)
	if err != nil {
		return entity.Clothing{}, fmt.Errorf("GetClothing %w", err)
	}

	return clothing, nil
}

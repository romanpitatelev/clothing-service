package brandsservice

import (
	"context"
	"fmt"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type brandsStore interface {
	ListBrands(ctx context.Context, req entity.ListRequest) ([]entity.Brand, error)
	ListPreferredBrands(ctx context.Context, userID entity.UserID) ([]entity.Brand, error)
	SetPreferredBrands(ctx context.Context, userID entity.UserID, brandIDs []entity.BrandID) error
	InsertRequestedBrand(ctx context.Context, req entity.RequestedBrand) error
}

type Service struct {
	brandsStore brandsStore
}

func New(brandsStore brandsStore) *Service {
	return &Service{
		brandsStore: brandsStore,
	}
}

func (s *Service) ListBrands(ctx context.Context, req entity.ListRequest) ([]entity.Brand, error) {
	brands, err := s.brandsStore.ListBrands(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListBrands %w", err)
	}

	return brands, nil
}

func (s *Service) ListPreferredBrands(ctx context.Context, userID entity.UserID) ([]entity.Brand, error) {
	brands, err := s.brandsStore.ListPreferredBrands(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ListPreferredBrands %w", err)
	}

	return brands, nil
}

func (s *Service) SetPreferredBrands(ctx context.Context, userID entity.UserID, brandIDs []entity.BrandID) error {
	if err := s.brandsStore.SetPreferredBrands(ctx, userID, brandIDs); err != nil {
		return fmt.Errorf("SetPreferredBrands %w", err)
	}

	return nil
}

func (s *Service) InsertRequestedBrand(ctx context.Context, req entity.RequestedBrand) error {
	if err := s.brandsStore.InsertRequestedBrand(ctx, req); err != nil {
		return fmt.Errorf("InsertRequestedBrand %w", err)
	}

	return nil
}

package clotheshandler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest/common"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type clothesService interface {
	GetClothing(ctx context.Context, id entity.ClothingID) (entity.Clothing, error)
}

type Handler struct {
	clothingService clothesService
}

func New(clothingService clothesService) *Handler {
	return &Handler{
		clothingService: clothingService,
	}
}

func (s *Handler) GetClothing(w http.ResponseWriter, r *http.Request) {
	clothingID, err := uuid.Parse(chi.URLParam(r, "clothingId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	clothing, err := s.clothingService.GetClothing(r.Context(), entity.ClothingID(clothingID))
	if err != nil {
		common.ErrorResponse(w, "error getting clothing", err)

		return
	}

	common.OkResponse(w, http.StatusOK, clothing)
}

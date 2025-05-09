package clotheshandler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest/common"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type productsService interface {
	GetVariant(ctx context.Context, id entity.VariantID) (entity.Variant, error)
}

type Handler struct {
	productsService productsService
}

func New(productsService productsService) *Handler {
	return &Handler{
		productsService: productsService,
	}
}

func (s *Handler) GetClothing(w http.ResponseWriter, r *http.Request) {
	clothingID, err := uuid.Parse(chi.URLParam(r, "clothingId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	clothing, err := s.productsService.GetVariant(r.Context(), entity.VariantID(clothingID))
	if err != nil {
		common.ErrorResponse(w, "error getting clothing", err)

		return
	}

	common.OkResponse(w, http.StatusOK, clothing)
}

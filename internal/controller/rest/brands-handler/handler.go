package brandshandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/romanpitatelev/clothing-service/internal/controller/rest/common"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type brandsService interface {
	ListBrands(ctx context.Context, req entity.ListRequest) ([]entity.Brand, error)
	ListPreferredBrands(ctx context.Context, userID entity.UserID) ([]entity.Brand, error)
	SetPreferredBrands(ctx context.Context, userID entity.UserID, brandIDs []entity.BrandID) error
	InsertRequestedBrand(ctx context.Context, req entity.RequestedBrand) error
}

type Handler struct {
	brandsService brandsService
}

func New(brandsService brandsService) *Handler {
	return &Handler{
		brandsService: brandsService,
	}
}

func (h *Handler) ListBrands(w http.ResponseWriter, r *http.Request) {
	req := common.GetListRequest(r)

	brands, err := h.brandsService.ListBrands(r.Context(), req)
	if err != nil {
		common.ErrorResponse(w, "error listing brands", err)

		return
	}

	common.OkResponse(w, http.StatusOK, brands)
}

func (h *Handler) SetPreferredBrands(w http.ResponseWriter, r *http.Request) {
	session := common.UserSession(r)

	var brandIDs []entity.BrandID

	if err := json.NewDecoder(r.Body).Decode(&brandIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if err := h.brandsService.SetPreferredBrands(r.Context(), session.UserID, brandIDs); err != nil {
		common.ErrorResponse(w, "error setting preferred brands", err)

		return
	}

	common.OkResponse(w, http.StatusOK, nil)
}

func (h *Handler) ListPreferredBrands(w http.ResponseWriter, r *http.Request) {
	session := common.UserSession(r)

	brands, err := h.brandsService.ListPreferredBrands(r.Context(), session.UserID)
	if err != nil {
		common.ErrorResponse(w, "error listing preferred brands", err)

		return
	}

	common.OkResponse(w, http.StatusOK, brands)
}

func (h *Handler) RequestBrand(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestedBrand

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	req.UserID = common.UserSession(r).UserID

	if err := h.brandsService.InsertRequestedBrand(r.Context(), req); err != nil {
		common.ErrorResponse(w, "error inserting brand", err)

		return
	}

	common.OkResponse(w, http.StatusCreated, nil)
}

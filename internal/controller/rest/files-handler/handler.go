package fileshandler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest/common"
	"github.com/rs/zerolog/log"
)

type filesService interface {
	GetFile(fileName string) (io.ReadCloser, string, error)
}

type Handler struct {
	filesService filesService
}

func New(filesService filesService) *Handler {
	return &Handler{
		filesService: filesService,
	}
}

func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "imageName")

	file, contentType, err := h.filesService.GetFile(name)
	if err != nil {
		common.ErrorResponse(w, "error deleting file", err)

		return
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Warn().Err(err).Msg("error closing file")
		}
	}()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	w.WriteHeader(http.StatusOK)

	if _, err = io.Copy(w, file); err != nil {
		log.Warn().Err(err).Msg("error copying file")
	}
}

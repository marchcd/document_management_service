package handlers

import (
	"net/http"
	"strconv"

	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/pkg"
	"github.com/marchcd/kai/internal/service"
)

type RegistryHandler struct {
	srv *service.DocumentService
}

func NewRegistryHandler(service *service.DocumentService) *RegistryHandler {
	return &RegistryHandler{srv: service}
}

func (h *RegistryHandler) PreviewRegistry(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "no from or to param",
		})
		return
	}

	rows, err := h.srv.GetRegistryRows(r.Context(), from, to)
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if rows == nil {
		rows = []models.RegistryRow{}
	}

	pkg.SendJSON(w, http.StatusOK, rows)
}

func (h *RegistryHandler) DownloadRegistry(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "no from or to param",
		})
		return
	}

	pdfBytes, err := h.srv.DownloadRegistry(r.Context(), from, to)
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	filename := "registry_" + from + "_" + to + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

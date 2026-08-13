package handlers

import (
	"net/http"
	"strings"

	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/pkg"
	"github.com/marchcd/kai/internal/service"
)

type StatusHandler struct {
	docService *service.DocumentService
}

func NewStatusHandler(s *service.DocumentService) *StatusHandler {
	return &StatusHandler{docService: s}
}

func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	studentCard := strings.TrimPrefix(r.URL.Path, "/api/status/")
	studentCard = strings.TrimSpace(studentCard)

	if studentCard == "" {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "student_card is required",
		})
		return
	}

	status, err := h.docService.GetStatus(r.Context(), studentCard)
	if err != nil {
		pkg.SendJSON(w, http.StatusNotFound, models.ErrorResponse{
			Error: "document not found",
		})
		return
	}

	pkg.SendJSON(w, http.StatusOK, status)
}

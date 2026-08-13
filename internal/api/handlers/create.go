package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/pkg"
	"github.com/marchcd/kai/internal/service"
)

type CreatePDFHandler struct {
	pdfService *service.DocumentService
}

func NewCreatePDFHandler(s *service.DocumentService) *CreatePDFHandler {
	return &CreatePDFHandler{pdfService: s}
}

func (h *CreatePDFHandler) CreatePDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "Request error",
		})
		return
	}

	resp, err := h.pdfService.CreatePDF(r.Context(), &user)
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	pkg.SendJSON(w, http.StatusOK, resp)
}

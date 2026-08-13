package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/pkg"
	"github.com/marchcd/kai/internal/service"
)

type RequestHandler struct {
	pdfService *service.DocumentService
}

func NewRequestsHandler(s *service.DocumentService) *RequestHandler {
	return &RequestHandler{pdfService: s}
}

func (h *RequestHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	items, err := h.pdfService.GetRequests(r.Context())
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if items == nil {
		items = []models.RequestListItem{}
	}

	pkg.SendJSON(w, http.StatusOK, items)
}

func (h *RequestHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	issueNumber, ok := parseIssueNumber(w, r.URL.Path, "/api/pdf/")
	if !ok {
		return
	}

	approvedAt := r.URL.Query().Get("approved_at")

	docBytes, err := h.pdfService.DownloadPDF(r.Context(), issueNumber, approvedAt)
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", "inline; filename=\"spravka_"+strconv.Itoa(issueNumber)+".docx\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(docBytes)))
	// w.Header().Set("Content-Type", "application/pdf")
	// w.Header().Set("Content-Disposition", "attachment; filename=\"spravka_"+strconv.Itoa(issueNumber)+".pdf\"")
	// w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(docBytes)
}

func (h *RequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
	issueNumber, ok := parseIssueNumber(w, r.URL.Path, "/api/requests/")
	if !ok {
		return
	}

	if err := h.pdfService.Approve(r.Context(), issueNumber); err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	pkg.SendJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (h *RequestHandler) Reject(w http.ResponseWriter, r *http.Request) {
	issueNumber, ok := parseIssueNumber(w, r.URL.Path, "/api/requests/")
	if !ok {
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	if err := h.pdfService.Reject(r.Context(), issueNumber, body.Reason); err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	pkg.SendJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (h *RegistryHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	issueNumber, ok := parseIssueNumber(w, r.URL.Path, "/api/requests/")
	if !ok {
		return
	}
	detail, err := h.srv.GetDocumentDetail(r.Context(), issueNumber)
	if err != nil {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	pkg.SendJSON(w, http.StatusOK, detail)
}

func (h *RegistryHandler) UpdateDetail(w http.ResponseWriter, r *http.Request) {
	issueNumber, ok := parseIssueNumber(w, r.URL.Path, "/api/requests/")
	if !ok {
		return
	}

	var body models.DocumentDetail
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	if err := h.srv.UpdateDocument(r.Context(), issueNumber, body); err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	pkg.SendJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func parseIssueNumber(w http.ResponseWriter, path, prefix string) (int, bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "issue_number is required",
		})
		return 0, false
	}

	n, err := strconv.Atoi(parts[0])
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "invalid issue_number",
		})
		return 0, false
	}

	return n, true
}

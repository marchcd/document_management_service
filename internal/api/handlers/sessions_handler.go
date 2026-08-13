package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/pkg"
	"github.com/marchcd/kai/internal/service"
)

type SessionsHandler struct {
	srv      *service.DocumentService
	filePath string
}

func NewSessionsHandler(s *service.DocumentService, filePath string) *SessionsHandler {
	return &SessionsHandler{
		srv:      s,
		filePath: filePath,
	}
}

func (h *SessionsHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(h.filePath)
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "Can't open sessions file",
		})
		return
	}
	defer file.Close()

	var sessions []models.DirectionJSON
	if err := json.NewDecoder(file).Decode(&sessions); err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "Error reading session file",
		})
		return
	}

	pkg.SendJSON(w, http.StatusOK, sessions)
}

func (h *SessionsHandler) SaveSessions(w http.ResponseWriter, r *http.Request) {
	var sessions []models.DirectionJSON
	if err := json.NewDecoder(r.Body).Decode(&sessions); err != nil {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "Bad request format",
		})
		return
	}

	if len(sessions) == 0 {
		pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "List of direction can't be empty",
		})
		return
	}

	for _, d := range sessions {
		for _, g := range d.Groups {
			if g.StartDate == "" || g.EndDate == "" {
				pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "All groups should have start date and end date"})
				return
			}

			if g.StartDate >= g.EndDate {
				pkg.SendJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "Start date >= End date"})
				return
			}
		}
	}

	data, err := json.MarshalIndent(sessions, "", " ")
	if err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "Serialization error",
		})
		return
	}

	if err := os.WriteFile(h.filePath, data, 0o644); err != nil {
		pkg.SendJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "Can't save the session file",
		})
		return
	}

	h.srv.ReloadSessions(sessions)
	pkg.SendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

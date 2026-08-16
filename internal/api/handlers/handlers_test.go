package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	mock.Mock
}

func (m *stubRepo) CreatePDF(ctx context.Context, data models.DataDocument) (*models.DocumentResponse, error) {
	args := m.Called(ctx, data)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.DocumentResponse), args.Error(1)
}

func (m *stubRepo) GetRequests(ctx context.Context) ([]models.RequestListItem, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].([]models.RequestListItem), args.Error(1)
}

func (m *stubRepo) GetDocumentByIssueNumber(ctx context.Context, issueNumber int) (*models.DataDocument, error) {
	args := m.Called(ctx, issueNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.DataDocument), args.Error(1)
}

func (m *stubRepo) SetStatus(ctx context.Context, issueNumber int, status string) error {
	args := m.Called(ctx, issueNumber, status)

	return args.Error(0)
}

func (m *stubRepo) SetRejected(ctx context.Context, issueNumber int, reason string) error {
	args := m.Called(ctx, issueNumber, reason)
	return args.Error(0)
}

func (m *stubRepo) GetRegistry(ctx context.Context, from, to string) ([]models.RegistryRow, error) {
	args := m.Called(ctx, from, to)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].([]models.RegistryRow), args.Error(1)
}

func (m *stubRepo) GetStatusByStudentCard(ctx context.Context, studentCard string) (*models.StatusResponse, error) {
	args := m.Called(ctx, studentCard)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.StatusResponse), args.Error(1)
}

func (m *stubRepo) GetDocumentDetail(ctx context.Context, issueNumber int) (*models.DocumentDetail, error) {
	args := m.Called(ctx, issueNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.DocumentDetail), args.Error(1)
}

func (m *stubRepo) UpdateDocument(ctx context.Context, issueNumber int, d models.DocumentDetail) error {
	args := m.Called(ctx, issueNumber, d)
	return args.Error(0)
}

func TestStatusHandler_EmptyStudentCard(t *testing.T) {
	h := NewStatusHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/status/", nil)
	rec := httptest.NewRecorder()

	h.GetStatus(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "student_card is required")
}

func TestSessionHandler_GetSessions(t *testing.T) {
	// Setup
	tests := []struct {
		name        string
		fileContent string
		fileExists  bool
		wantStatus  int
	}{
		{
			name: "успешное чтение",
			fileContent: `[
				{
				"direction": "09.03.01 Информатика и вычислительная техника",
				"groups": [
				{
					"group": 24174,
					"start_date": "2026-06-01",
					"end_date": "2026-06-19"
   				}]}]`,
			fileExists: true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "файл не существует",
			fileExists: false,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "битый json в файле",
			fileContent: "{битый json",
			fileExists:  true,
			wantStatus:  http.StatusInternalServerError,
		},
	}

	// Act and assertion
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.TempDir() создаёт временную папку и сам удаляет
			// её после завершения теста — не нужен явный cleanup
			dir := t.TempDir()
			filePath := filepath.Join(dir, "sessions.json")

			if tt.fileExists {
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0o644)
				require.NoError(t, err)
			}

			h := NewSessionsHandler(nil, filePath)
			req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
			rec := httptest.NewRecorder()

			h.GetSessions(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestSessionHandler_SaveSessions(t *testing.T) {
	// Setup
	validBody := `[
				{
				"direction": "09.03.01 Информатика и вычислительная техника",
				"groups": [
				{
					"group": 24174,
					"start_date": "2026-06-01",
					"end_date": "2026-06-19"
   				}]}]`
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "успешное сохранение",
			body:       validBody,
			wantStatus: http.StatusOK,
		},
		{
			name:       "ошибка json",
			body:       `не json`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "Bad request format",
		},
		{
			name:       "пустой список JSON",
			body:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "List of direction can't be empty",
		},
		{
			name: "startDate >= endDate",
			body: `[
				{
				"direction": "09.03.01 Информатика и вычислительная техника",
				"groups": [
				{
					"group": 24174,
					"start_date": "2026-06-19",
					"end_date": "2026-06-01"
   				}]}]`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "Start date bigger than End date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirTemp := t.TempDir()
			filePath := filepath.Join(dirTemp, "sessions.json")

			svc := service.NewDocumentService(&stubRepo{}, nil)

			h := NewSessionsHandler(svc, filePath)
			req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.SaveSessions(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantErr != "" {
				assert.Contains(t, rec.Body.String(), tt.wantErr)
			}
		})
	}
}

func TestRegistryHandler_PreviewRegistry(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		setupMock  func(m *stubRepo)
	}{
		{
			name:       "успешное чтение",
			path:       "/api/registry/preview?from=2026-01-01&to=2026-01-10",
			wantStatus: http.StatusOK,
			setupMock: func(m *stubRepo) {
				m.On("GetRegistry", mock.Anything, "2026-01-01", "2026-01-10").Return([]models.RegistryRow{}, nil)
			},
		},
		{
			name:       "ошибка параметров",
			path:       "/api/registry/preview",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ошибка параметров (только from)",
			path:       "/api/registry/preview?from=2026-01-01",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := new(stubRepo)
			if tt.setupMock != nil {
				tt.setupMock(repoMock)
			}

			srv := service.NewDocumentService(repoMock, nil)
			h := NewRegistryHandler(srv)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			h.PreviewRegistry(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			repoMock.AssertExpectations(t)
		})
	}
}

func TestRegistryHandler_DownloadRegistry(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "ошибка параметра from",
			path:       "/api/registry?from=2026-01-10&to=",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := new(stubRepo)

			srv := service.NewDocumentService(repoMock, nil)
			h := NewRegistryHandler(srv)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			h.DownloadRegistry(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

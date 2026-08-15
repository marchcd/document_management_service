package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/marchcd/kai/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type DocumentServiceMock struct {
	mock.Mock
}

func (m *DocumentServiceMock) CreatePDF(ctx context.Context, data models.DataDocument) (*models.DocumentResponse, error) {
	args := m.Called(ctx, data)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.DocumentResponse), args.Error(1)
}

func (m *DocumentServiceMock) GetRequests(ctx context.Context) ([]models.RequestListItem, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].([]models.RequestListItem), args.Error(1)
}

func (m *DocumentServiceMock) GetDocumentByIssueNumber(ctx context.Context, issueNumber int) (*models.DataDocument, error) {
	args := m.Called(ctx, issueNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.DataDocument), args.Error(1)
}

func (m *DocumentServiceMock) SetStatus(ctx context.Context, issueNumber int, status string) error {
	args := m.Called(ctx, issueNumber, status)

	return args.Error(0)
}

func (m *DocumentServiceMock) SetRejected(ctx context.Context, issueNumber int, reason string) error {
	args := m.Called(ctx, issueNumber, reason)
	return args.Error(0)
}

func (m *DocumentServiceMock) GetRegistry(ctx context.Context, from, to string) ([]models.RegistryRow, error) {
	args := m.Called(ctx, from, to)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].([]models.RegistryRow), args.Error(1)
}

func (m *DocumentServiceMock) GetStatusByStudentCard(ctx context.Context, studentCard string) (*models.StatusResponse, error) {
	args := m.Called(ctx, studentCard)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.StatusResponse), args.Error(1)
}

func (m *DocumentServiceMock) GetDocumentDetail(ctx context.Context, issueNumber int) (*models.DocumentDetail, error) {
	args := m.Called(ctx, issueNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args[0].(*models.DocumentDetail), args.Error(1)
}

func (m *DocumentServiceMock) UpdateDocument(ctx context.Context, issueNumber int, d models.DocumentDetail) error {
	args := m.Called(ctx, issueNumber, d)
	return args.Error(0)
}

func TestCreatePDF(t *testing.T) {
	// Setup
	sessionsData := []models.DirectionJSON{
		{
			Direction: "09.03.01 Информатика и вычислительная техника",
			Groups: []models.GroupJSON{
				{Group: 24174, StartDate: "2026-06-01", EndDate: "2026-06-19"},
			},
		},
	}

	tests := []struct {
		name      string
		input     *models.UserRequest
		mockSetup func(m *DocumentServiceMock)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "успешное создание документа",
			input: &models.UserRequest{
				StudentCard:   "12345",
				FullTitle:     "Иванов Иван Иванович",
				EmployerName:  "ООО Компания",
				EducationForm: "заочная",
				Course:        1,
				StudyGroup:    24174,
			},
			mockSetup: func(m *DocumentServiceMock) {
				m.On("CreatePDF", mock.Anything, mock.AnythingOfType("models.DataDocument")).Return(&models.DocumentResponse{IssueNumber: 1}, nil)
			},
			wantErr: false,
		},
		{
			name: "группа не найдена",
			input: &models.UserRequest{
				StudentCard:   "12345",
				FullTitle:     "Иванов Иван Иванович",
				EducationForm: "очная",
				Course:        2,
				StudyGroup:    99999,
				EmployerName:  "ООО Ромашка",
			},
			mockSetup: func(m *DocumentServiceMock) {
			},
			wantErr: true,
			errMsg:  "Группа не найдена",
		},
		{
			name: "репозиторий вернул ошибку",
			input: &models.UserRequest{
				StudentCard:   "12345",
				FullTitle:     "Иванов Иван Иванович",
				EducationForm: "очная",
				Course:        2,
				StudyGroup:    24174,
				EmployerName:  "ООО Ромашка",
			},
			mockSetup: func(m *DocumentServiceMock) {
				m.On("CreatePDF", mock.Anything, mock.AnythingOfType("models.DataDocument")).Return(nil, errors.New("Вы уже создали заявку. Ожидайте!"))
			},
			wantErr: true,
			errMsg:  "Вы уже создали заявку",
		},
	}

	// Execution and assertion
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(DocumentServiceMock)
			tt.mockSetup(repo)

			svc := NewDocumentService(repo, sessionsData)
			resp, err := svc.CreatePDF(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, 1, resp.IssueNumber)
				assert.Equal(t, 1, resp.IssueNumber)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestDownloadPDF(t *testing.T) {
	// Setup
	raw := &models.DataDocument{
		StudentType: models.Student{
			StudentCard: "12345",
			FullTitle:   "Иванов Иван Иванович",
			Course:      2,
		},
		DocumentType: models.Document{
			Number:       7,
			PeriodStart:  time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:    time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC),
			Duration:     19,
			EmployerName: "ООО Компания",
		},
	}

	tests := []struct {
		name       string
		approvedAt string
		wantDay    string
		wantMonth  string
	}{
		{
			name:       "валиданая дата утверждения",
			approvedAt: "15.08.2026",
			wantDay:    "15",
			wantMonth:  "августа",
		},
		{
			name:       "невалидная дата - используется time.Now()",
			approvedAt: "не дата",
		},
	}

	// Execution and assertion
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDocumentData(raw, tt.approvedAt)

			require.NotNil(t, got)
			assert.Equal(t, "0007", got.Issue.Number)
			assert.Equal(t, "01", got.Period.StartDay)
			assert.Equal(t, "июня", got.Period.StartMonth)
			assert.Equal(t, 19, got.Period.Duration)

			if tt.wantDay != "" {
				assert.Equal(t, tt.wantDay, got.Issue.Day)
				assert.Equal(t, tt.wantMonth, got.Issue.Month)
			}
		})
	}
}

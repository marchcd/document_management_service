package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/pkg"
)

type DocumentRepository interface {
	CreatePDF(ctx context.Context, data models.DataDocument) (*models.DocumentResponse, error)
	GetRequests(ctx context.Context) ([]models.RequestListItem, error)
	GetDocumentByIssueNumber(ctx context.Context, issueNumber int) (*models.DataDocument, error)
	SetStatus(ctx context.Context, issueNumber int, status string) error
	SetRejected(ctx context.Context, issueNumber int, reason string) error
	GetRegistry(ctx context.Context, from, to string) ([]models.RegistryRow, error)
	GetStatusByStudentCard(ctx context.Context, studentCard string) (*models.StatusResponse, error)
	GetDocumentDetail(ctx context.Context, issueNumber int) (*models.DocumentDetail, error)
	UpdateDocument(ctx context.Context, issueNumber int, d models.DocumentDetail) error
}

type DocumentService struct {
	repo        DocumentRepository
	sessionsMap map[int]models.GroupInfo
}

func NewDocumentService(r DocumentRepository, sessionsData []models.DirectionJSON) *DocumentService {
	sMap := make(map[int]models.GroupInfo)

	for _, d := range sessionsData {
		for _, g := range d.Groups {
			sMap[g.Group] = models.GroupInfo{
				Direction: d.Direction,
				StartDate: g.StartDate,
				EndDate:   g.EndDate,
			}
		}
	}

	return &DocumentService{repo: r, sessionsMap: sMap}
}

func (s *DocumentService) CreatePDF(ctx context.Context, user *models.UserRequest) (*models.DocumentResponse, error) {
	// I have to fill DataDocument with values to save it in db
	info, ok := s.sessionsMap[user.StudyGroup]
	if !ok {
		return nil, fmt.Errorf("Группа не найдена")
	}

	// To dative
	dativeName := pkg.InflectDative(user.FullTitle)

	layout := "2006-01-02"

	start, err := time.Parse(layout, info.StartDate)
	if err != nil {
		return nil, fmt.Errorf("Error startDate parsing")
	}

	end, err := time.Parse(layout, info.EndDate)
	if err != nil {
		return nil, fmt.Errorf("Error endDate parsing")
	}

	diff := end.Sub(start)
	days := int(diff.Hours()/24) + 1

	data := models.DataDocument{
		StudentType: models.Student{
			StudentCard:         user.StudentCard,
			FullTitle:           dativeName,
			FullTitleNominative: user.FullTitle,
			EducationForm:       user.EducationForm,
			Course:              user.Course,
			Group:               user.StudyGroup,
			Direction:           info.Direction,
		},
		DocumentType: models.Document{
			PeriodStart:  start,
			PeriodEnd:    end,
			Duration:     days,
			EmployerName: user.EmployerName,
		},
	}

	return s.repo.CreatePDF(ctx, data)
}

func (s *DocumentService) GetRequests(ctx context.Context) ([]models.RequestListItem, error) {
	return s.repo.GetRequests(ctx)
}

func (s *DocumentService) DownloadPDF(ctx context.Context, issueNumber int, approvedAt string) ([]byte, error) {
	raw, err := s.repo.GetDocumentByIssueNumber(ctx, issueNumber)
	if err != nil {
		return nil, err
	}

	russianMonths := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря",
	}

	var issueTime time.Time
	if approvedAt != "" {
		parsed, err := time.Parse("02.01.2006", approvedAt)
		if err == nil {
			issueTime = parsed
		} else {
			issueTime = time.Now()
		}
	} else {
		issueTime = time.Now()
	}

	start := raw.DocumentType.PeriodStart
	end := raw.DocumentType.PeriodEnd

	docData := &models.DocumentData{
		Issue: models.IssueInfo{
			Number:    fmt.Sprintf("%04d", raw.DocumentType.Number),
			Day:       fmt.Sprintf("%02d", issueTime.Day()),
			Month:     russianMonths[issueTime.Month()-1],
			YearShort: fmt.Sprintf("%02d", issueTime.Year()%100),
		},
		Employer: models.EmployerInfo{
			Name: raw.DocumentType.EmployerName,
		},
		Student: models.StudentInfo{
			StudentCard:         raw.StudentType.StudentCard,
			FullTitle:           raw.StudentType.FullTitle,
			FullTitleNominative: raw.StudentType.FullTitleNominative,
			EducationForm:       raw.StudentType.EducationForm,
			Course:              raw.StudentType.Course,
			Specialty:           raw.StudentType.Direction,
		},
		Period: models.PeriodInfo{
			StartDay:   fmt.Sprintf("%02d", start.Day()),
			StartMonth: russianMonths[start.Month()-1],
			StartYear:  strconv.Itoa(start.Year()),
			EndDay:     fmt.Sprintf("%02d", end.Day()),
			EndMonth:   russianMonths[end.Month()-1],
			EndYear:    strconv.Itoa(end.Year()),
			Duration:   raw.DocumentType.Duration,
		},
	}

	// return pkg.CreatePDF(ctx, docData)
	return pkg.CreateDocx(docData)
}

func (s *DocumentService) Approve(ctx context.Context, issueNumber int) error {
	return s.repo.SetStatus(ctx, issueNumber, "approved")
}

func (s *DocumentService) Reject(ctx context.Context, issueNumber int, reason string) error {
	return s.repo.SetRejected(ctx, issueNumber, reason)
}

func (s *DocumentService) GetStatus(ctx context.Context, studentCard string) (*models.StatusResponse, error) {
	return s.repo.GetStatusByStudentCard(ctx, studentCard)
}

func (s *DocumentService) GetDocumentDetail(ctx context.Context, issueNumber int) (*models.DocumentDetail, error) {
	return s.repo.GetDocumentDetail(ctx, issueNumber)
}

func (s *DocumentService) UpdateDocument(ctx context.Context, issueNumber int, data models.DocumentDetail) error {
	return s.repo.UpdateDocument(ctx, issueNumber, data)
}

func (s *DocumentService) ReloadSessions(sessionsData []models.DirectionJSON) {
	sMap := make(map[int]models.GroupInfo)
	for _, d := range sessionsData {
		for _, g := range d.Groups {
			sMap[g.Group] = models.GroupInfo{
				Direction: d.Direction,
				StartDate: g.StartDate,
				EndDate:   g.EndDate,
			}
		}
	}
	s.sessionsMap = sMap
}

func (s *DocumentService) GetRegistryRows(ctx context.Context, from, to string) ([]models.RegistryRow, error) {
	return s.repo.GetRegistry(ctx, from, to)
}

func (s *DocumentService) DownloadRegistry(ctx context.Context, from, to string) ([]byte, error) {
	rows, err := s.repo.GetRegistry(ctx, from, to)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	russianMonths := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря",
	}

	fmtDate := func(iso string) string {
		t, err := time.Parse("2006-01-02", iso)
		if err != nil {
			return iso
		}
		return fmt.Sprintf("%02d %s %d", t.Day(), russianMonths[t.Month()-1], t.Year())
	}

	data := &models.RegistryData{
		PeriodFrom:  fmtDate(from),
		PeriodTo:    fmtDate(to),
		TotalCount:  len(rows),
		GeneratedAt: fmt.Sprintf("%02d.%02d.%d", now.Day(), int(now.Month()), now.Year()),
		Rows:        rows,
	}

	return pkg.CreateRegistryPDF(ctx, data)
}

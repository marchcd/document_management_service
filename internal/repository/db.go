package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marchcd/kai/internal/models"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreatePDF(ctx context.Context, data models.DataDocument) (*models.DocumentResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	upsertStudent := `
			INSERT INTO students (student_card, full_title, full_title_nominative, education_form, course, study_group, direction)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (student_card) DO UPDATE
				SET full_title = EXCLUDED.full_title,
					full_title_nominative = EXCLUDED.full_title_nominative,
					education_form = EXCLUDED.education_form,
					course = EXCLUDED.course,
					study_group = EXCLUDED.study_group,
					direction = EXCLUDED.direction
			RETURNING id;
	`
	var studentID string

	err = tx.QueryRow(ctx, upsertStudent,
		data.StudentType.StudentCard, data.StudentType.FullTitle,
		data.StudentType.FullTitleNominative,
		data.StudentType.EducationForm, data.StudentType.Course,
		data.StudentType.Group, data.StudentType.Direction,
	).Scan(&studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert student: %w", err)
	}

	eligibilityCheck := `
		SELECT
			CASE
				WHEN EXISTS (
					SELECT 1 FROM documents
					WHERE student_id = $1 AND status = 'pending'
				) THEN 'Вы уже создали заявку. Ожидайте!'
				WHEN EXISTS (
					SELECT 1 FROM documents
					WHERE student_id = $1 
						AND status = 'approved' 
						AND period_start = $2
						AND period_end = $3
				) THEN 'Справка на данный период уже была выдана!'
				
				ELSE ''
			END;
	`

	var reason string
	if err = tx.QueryRow(ctx, eligibilityCheck, studentID, data.DocumentType.PeriodStart, data.DocumentType.PeriodEnd).Scan(&reason); err != nil {
		return nil, fmt.Errorf("eligibility check failed: %w", err)
	}

	if reason != "" {
		return nil, fmt.Errorf("%s", reason)
	}

	insertDoc := `
		INSERT INTO documents (student_id, period_start, period_end, duration, employer_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING issue_number;
	`

	var newDoc models.DocumentResponse
	if err := tx.QueryRow(ctx, insertDoc, studentID,
		data.DocumentType.PeriodStart, data.DocumentType.PeriodEnd,
		data.DocumentType.Duration, data.DocumentType.EmployerName,
	).Scan(&newDoc.IssueNumber); err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &newDoc, nil
}

func (s *PostgresStore) GetRequests(ctx context.Context) ([]models.RequestListItem, error) {
	query := `
		SELECT
			d.issue_number,
			s.full_title,
			s.student_card,
			s.study_group,
			TO_CHAR(d.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
		FROM documents d
		JOIN students s ON s.id = d.student_id
		WHERE d.status = 'pending'
		ORDER BY d.created_at DESC;
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var items []models.RequestListItem
	for rows.Next() {
		var item models.RequestListItem
		if err := rows.Scan(&item.IssueNumber, &item.FullTitle, &item.StudentCard, &item.StudyGroup, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *PostgresStore) GetDocumentByIssueNumber(ctx context.Context, issueNumber int) (*models.DataDocument, error) {
	query := `
		SELECT
			s.student_card,
			s.full_title,
			s.full_title_nominative,
			s.education_form,
			s.course,
			s.study_group,
			s.direction,
			d.issue_number,
			d.period_start,
			d.period_end,
			d.duration,
			d.employer_name,
			d.created_at
		FROM documents d
		JOIN students s ON s.id = d.student_id
		WHERE d.issue_number = $1;
	`

	var doc models.DataDocument
	err := s.pool.QueryRow(ctx, query, issueNumber).Scan(
		&doc.StudentType.StudentCard,
		&doc.StudentType.FullTitle,
		&doc.StudentType.FullTitleNominative,
		&doc.StudentType.EducationForm,
		&doc.StudentType.Course,
		&doc.StudentType.Group,
		&doc.StudentType.Direction,
		&doc.DocumentType.Number,
		&doc.DocumentType.PeriodStart,
		&doc.DocumentType.PeriodEnd,
		&doc.DocumentType.Duration,
		&doc.DocumentType.EmployerName,
		&doc.DocumentType.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	return &doc, nil
}

func (s *PostgresStore) SetStatus(ctx context.Context, issueNumber int, status string) error {
	query := `
		UPDATE documents 
		SET status = $1,
			approved_at = CASE WHEN $2 = 'approved' THEN NOW() ELSE approved_at END
		WHERE issue_number = $3`

	tag, err := s.pool.Exec(ctx, query, status, status, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to set status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("document %d not found", issueNumber)
	}

	return nil
}

func (s *PostgresStore) SetRejected(ctx context.Context, issueNumber int, reason string) error {
	query := `UPDATE documents SET status = 'rejected', reject_reason = $1 WHERE issue_number = $2`

	tag, err := s.pool.Exec(ctx, query, reason, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to reject document: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("document %d not found", issueNumber)
	}

	return nil
}

func (s *PostgresStore) GetRegistry(ctx context.Context, from, to string) ([]models.RegistryRow, error) {
	query := `
		SELECT
			d.issue_number,
			s.full_title_nominative,
			s.study_group,
			s.direction,
			COALESCE(d.employer_name, '') AS employer_name,
			TO_CHAR(d.approved_at AT TIME ZONE 'UTC', 'DD.MM.YYYY') AS issued_at
		FROM documents d
		JOIN students s ON s.id = d.student_id
		WHERE d.status = 'approved'
			AND d.approved_at::date >= $1::date
			AND d.approved_at::date <= $2::date
		ORDER BY d.approved_at ASC;
	`

	rows, err := s.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("registry query failed: %w", err)
	}
	defer rows.Close()

	var results []models.RegistryRow
	for rows.Next() {
		var r models.RegistryRow
		if err := rows.Scan(
			&r.IssueNumber, &r.FullTitleNominative, &r.StudyGroup, &r.Direction,
			&r.EmployerName, &r.IssuedAt,
		); err != nil {
			return nil, fmt.Errorf("registry scan failed: %w", err)
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

func (s *PostgresStore) GetStatusByStudentCard(ctx context.Context, studentCard string) (*models.StatusResponse, error) {
	query := `
		SELECT
			d.issue_number,
			s.full_title,
			d.status,
			COALESCE(d.reject_reason, '') AS reject_reason,
			TO_CHAR(d.period_start, 'DD.MM.YYYY') AS period_start,
			TO_CHAR(d.period_end, 'DD.MM.YYYY') AS period_end,
			TO_CHAR(d.created_at AT TIME ZONE 'UTC', 'DD.MM.YYYY') AS created_at
		FROM documents d
		JOIN students s ON s.id = d.student_id
		WHERE s.student_card = $1
		ORDER BY d.created_at DESC
		LIMIT 1;
	`

	var r models.StatusResponse
	err := s.pool.QueryRow(ctx, query, studentCard).Scan(&r.IssueNumber, &r.FullTitle, &r.Status, &r.RejectReason,
		&r.PeriodStart, &r.PeriodEnd, &r.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("not found")
	}

	return &r, nil
}

func (s *PostgresStore) GetDocumentDetail(ctx context.Context, issueNumber int) (*models.DocumentDetail, error) {
	query := `
		SELECT 
			d.issue_number,
			s.full_title,
			s.student_card,
			s.education_form,
			s.course,
			s.study_group,
			s.direction,
			COALESCE(d.employer_name, '') AS employer_name,
			TO_CHAR(d.period_start, 'YYYY-MM-DD') AS period_start,
			TO_CHAR(d.period_end, 'YYYY-MM-DD') AS period_end,
			d.duration,
			TO_CHAR(d.created_at AT TIME ZONE 'UTC', 'DD.MM.YYYY') AS created_at
		FROM documents d
		JOIN students s ON s.id = d.student_id
		WHERE d.issue_number = $1;
	`

	var r models.DocumentDetail
	err := s.pool.QueryRow(ctx, query, issueNumber).Scan(
		&r.IssueNumber, &r.FullTitle, &r.StudentCard, &r.EducationForm,
		&r.Course, &r.StudyGroup, &r.Direction, &r.EmployerName,
		&r.PeriodStart, &r.PeriodEnd, &r.Duration, &r.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	return &r, nil
}

func (s *PostgresStore) UpdateDocument(ctx context.Context, issueNumber int, d models.DocumentDetail) error {
	query := `
		WITH updated_docs AS (
				UPDATE documents
				SET employer_name = $1
				WHERE issue_number = $2
				RETURNING student_id
		)
		UPDATE students
		SET
				full_title = $3,
				education_form = $4,
				course = $5
		FROM updated_docs
		WHERE students.id = updated_docs.student_id;
	`

	_, err := s.pool.Exec(ctx, query,
		d.EmployerName, issueNumber, d.FullTitle, d.EducationForm, d.Course,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления документа: %w", err)
	}

	return err
}

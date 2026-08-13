package models

import "time"

type Student struct {
	ID          string `db:"id"`
	StudentCard string `db:"student_card"`
	// FullTitle - Dative name
	FullTitle string `db:"full_title"`
	// FullTitleNominative - Nominative name
	FullTitleNominative string `db:"full_title_nominative"`
	EducationForm       string `db:"education_form"`
	Course              int    `db:"course"`
	Group               int    `db:"study_group"`
	Direction           string `db:"direction"`
}

type Document struct {
	ID           string    `db:"id"`
	Number       int       `db:"issue_number"`
	PeriodStart  time.Time `db:"period_start"`
	PeriodEnd    time.Time `db:"period_end"`
	Duration     int       `db:"duration"`
	EmployerName string    `db:"employer_name"`
	CreatedAt    time.Time `db:"created_at"`
}

type DataDocument struct {
	StudentType  Student
	DocumentType Document
}

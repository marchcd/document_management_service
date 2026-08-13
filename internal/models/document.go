package models

type DocumentResponse struct {
	IssueNumber int `db:"issue_number" json:"issue_number"`
}

type DocumentData struct {
	Issue    IssueInfo
	Employer EmployerInfo
	Student  StudentInfo
	Period   PeriodInfo
}

type IssueInfo struct {
	Number    string
	Day       string
	Month     string
	YearShort string
}

type PeriodInfo struct {
	StartDay   string
	StartMonth string
	StartYear  string
	EndDay     string
	EndMonth   string
	EndYear    string
	Duration   int
}

type StudentInfo struct {
	StudentCard         string
	FullTitle           string
	FullTitleNominative string
	EducationForm       string
	Course              int
	Specialty           string
}

type EmployerInfo struct {
	Name string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type GroupInfo struct {
	Direction string
	StartDate string
	EndDate   string
}

type RegistryRow struct {
	IssueNumber         int    `json:"issue_number"`
	FullTitleNominative string `json:"full_title_nominative"`
	StudyGroup          int    `json:"study_group"`
	Direction           string `json:"direction"`
	EmployerName        string `json:"employer_name"`
	IssuedAt            string `json:"issued_at"`
}

type RegistryData struct {
	PeriodFrom  string
	PeriodTo    string
	TotalCount  int
	GeneratedAt string
	Rows        []RegistryRow
}

type StatusResponse struct {
	IssueNumber  int    `json:"issue_number"`
	FullTitle    string `json:"full_title"`
	Status       string `json:"status"`
	RejectReason string `json:"reject_reason"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
	CreatedAt    string `json:"created_at"`
}

type DocumentDetail struct {
	IssueNumber   int    `json:"issue_number"`
	FullTitle     string `json:"full_title"`
	StudentCard   string `json:"student_card"`
	EducationForm string `json:"education_form"`
	Course        int    `json:"course"`
	StudyGroup    int    `json:"study_group"`
	Direction     string `json:"direction"`
	EmployerName  string `json:"employer_name"`
	PeriodStart   string `json:"period_start"`
	PeriodEnd     string `json:"period_end"`
	Duration      int    `json:"duration"`
	CreatedAt     string `json:"created_at"`
}

package models

type UserRequest struct {
	EmployerName  string `json:"employer_name"`
	FullTitle     string `json:"full_title"`
	StudentCard   string `json:"student_card"`
	EducationForm string `json:"education_form"`
	Course        int    `json:"course"`
	StudyGroup    int    `json:"study_group"`
}

type DirectionJSON struct {
	Direction string      `json:"direction"`
	Groups    []GroupJSON `json:"groups"`
}

type GroupJSON struct {
	Group     int    `json:"group"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type RequestListItem struct {
	IssueNumber int    `json:"issue_number" db:"issue_number"`
	FullTitle   string `json:"full_title" db:"full_title"`
	StudentCard string `json:"student_card" db:"student_card"`
	StudyGroup  int    `json:"study_group" db:"study_group"`
	CreatedAt   string `json:"created_at" db:"created_at"`
}

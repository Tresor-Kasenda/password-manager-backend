package models

type ImportEntry struct {
	Title            string   `json:"title"`
	Website          *string  `json:"website"`
	Username         *string  `json:"username"`
	Password         string   `json:"password"`
	Notes            *string  `json:"notes"`
	Folder           *string  `json:"folder"`
	Favorite         bool     `json:"favorite"`
	Source           string   `json:"source"`
	ValidationIssues []string `json:"validation_issues,omitempty"`
}

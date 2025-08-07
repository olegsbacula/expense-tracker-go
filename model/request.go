package model

type Request struct {
	ID          string `json:"id"`
	Expenses    string `json:"expenses"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

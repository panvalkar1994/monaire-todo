package domain

import "time"

type Todo struct {
	ID        string
	Text      string
	DueDate   time.Time
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TodoPatch struct {
	Text      *string
	DueDate   *time.Time
	Completed *bool
}

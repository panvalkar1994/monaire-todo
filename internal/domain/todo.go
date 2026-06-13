package domain

import "time"

type Todo struct {
	ID        string
	Description string
	DueDate   time.Time
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TodoPatch struct {
	Description *string
	DueDate   *time.Time
	Completed *bool
}

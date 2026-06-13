package gormrepo

import "time"

type TodoModel struct {
	ID        string    `gorm:"column:id;primaryKey;size:36"`
	Text      string    `gorm:"column:text;size:500;not null"`
	DueDate   time.Time `gorm:"column:due_date;type:date;not null"`
	Completed bool      `gorm:"column:completed;not null;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TodoModel) TableName() string {
	return "todos"
}

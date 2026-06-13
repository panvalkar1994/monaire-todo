package gormrepo

import (
	"context"
	"errors"
	"time"

	"todo/internal/domain"
	"todo/internal/repository"

	"gorm.io/gorm"
)

type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) repository.TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	model := toModel(todo)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	*todo = fromModel(model)
	return nil
}

func (r *TodoRepository) GetByID(ctx context.Context, id string) (*domain.Todo, error) {
	var model TodoModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	todo := fromModel(model)
	return &todo, nil
}

func (r *TodoRepository) List(ctx context.Context, includeCompleted bool) ([]*domain.Todo, error) {
	q := r.db.WithContext(ctx).Model(&TodoModel{})
	if !includeCompleted {
		q = q.Where("completed = ?", false)
	}
	q = q.Order("due_date ASC")

	var models []TodoModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Todo, 0, len(models))
	for _, m := range models {
		t := fromModel(m)
		out = append(out, &t)
	}
	return out, nil
}

func (r *TodoRepository) Replace(ctx context.Context, todo *domain.Todo) error {
	model := toModel(todo)
	result := r.db.WithContext(ctx).Model(&TodoModel{}).Where("id = ?", todo.ID).Updates(map[string]interface{}{
		"description": model.Description,
		"due_date":   model.DueDate,
		"completed":  model.Completed,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNoChanges
	}
	return r.refresh(ctx, todo)
}

func (r *TodoRepository) Update(ctx context.Context, id string, patch domain.TodoPatch) error {
	updates := map[string]any{}
	if patch.Description != nil {
		updates["description"] = *patch.Description
	}
	if patch.DueDate != nil {
		updates["due_date"] = *patch.DueDate
	}
	if patch.Completed != nil {
		updates["completed"] = *patch.Completed
	}
	if len(updates) == 0 {
		return domain.ErrValidation
	}
	updates["updated_at"] = time.Now()

	result := r.db.WithContext(ctx).Model(&TodoModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TodoRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&TodoModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TodoRepository) refresh(ctx context.Context, todo *domain.Todo) error {
	updated, err := r.GetByID(ctx, todo.ID)
	if err != nil {
		return err
	}
	*todo = *updated
	return nil
}

func toModel(t *domain.Todo) TodoModel {
	return TodoModel{
		ID:          t.ID,
		Description: t.Description,
		DueDate:   t.DueDate,
		Completed: t.Completed,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func fromModel(m TodoModel) domain.Todo {
	return domain.Todo{
		ID:          m.ID,
		Description: m.Description,
		DueDate:   m.DueDate,
		Completed: m.Completed,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

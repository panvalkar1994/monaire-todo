package repository

import (
	"context"

	"todo/internal/domain"
)

type TodoRepository interface {
	Create(ctx context.Context, todo *domain.Todo) error
	GetByID(ctx context.Context, id string) (*domain.Todo, error)
	List(ctx context.Context, includeCompleted bool) ([]*domain.Todo, error)
	Replace(ctx context.Context, todo *domain.Todo) error
	Update(ctx context.Context, id string, patch domain.TodoPatch) error
	Delete(ctx context.Context, id string) error
}

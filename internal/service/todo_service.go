package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"todo/internal/domain"
	"todo/internal/repository"

	"github.com/google/uuid"
)

type TodoService struct {
	repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

type CreateInput struct {
	Description string
	DueDate     string
	Completed   *bool
}

type ReplaceInput struct {
	Description string
	DueDate     string
	Completed   bool
}

type PatchInput struct {
	Description *string
	DueDate     *string
	Completed   *bool
}

type ReplaceResult struct {
	Todo      *domain.Todo
	NoChanges bool
}

func (s *TodoService) Create(ctx context.Context, in CreateInput) (*domain.Todo, error) {
	description, err := validateDescription(in.Description)
	if err != nil {
		return nil, err
	}
	due, err := parseDate(in.DueDate)
	if err != nil {
		return nil, err
	}
	completed := false
	if in.Completed != nil {
		completed = *in.Completed
	}
	todo := &domain.Todo{
		ID:          uuid.NewString(),
		Description: description,
		DueDate:     due,
		Completed:   completed,
	}
	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, err
	}
	return todo, nil
}

func (s *TodoService) GetByID(ctx context.Context, id string) (*domain.Todo, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TodoService) List(ctx context.Context, includeCompleted bool) ([]*domain.Todo, error) {
	return s.repo.List(ctx, includeCompleted)
}

func (s *TodoService) Replace(ctx context.Context, id string, in ReplaceInput) (*ReplaceResult, error) {
	description, err := validateDescription(in.Description)
	if err != nil {
		return nil, err
	}
	due, err := parseDate(in.DueDate)
	if err != nil {
		return nil, err
	}

	todo := &domain.Todo{
		ID:          id,
		Description: description,
		DueDate:     due,
		Completed:   in.Completed,
	}

	_, err = s.repo.GetByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		if err := s.repo.Create(ctx, todo); err != nil {
			return nil, err
		}
		return &ReplaceResult{Todo: todo}, nil
	}
	if err != nil {
		return nil, err
	}

	if err := s.repo.Replace(ctx, todo); err != nil {
		if errors.Is(err, domain.ErrNoChanges) {
			existing, getErr := s.repo.GetByID(ctx, id)
			if getErr != nil {
				return nil, getErr
			}
			return &ReplaceResult{Todo: existing, NoChanges: true}, nil
		}
		return nil, err
	}
	return &ReplaceResult{Todo: todo}, nil
}

func (s *TodoService) Patch(ctx context.Context, id string, in PatchInput) (*domain.Todo, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return nil, err
	}
	patch := domain.TodoPatch{}
	if in.Description != nil {
		description, err := validateDescription(*in.Description)
		if err != nil {
			return nil, err
		}
		patch.Description = &description
	}
	if in.DueDate != nil {
		due, err := parseDate(*in.DueDate)
		if err != nil {
			return nil, err
		}
		patch.DueDate = &due
	}
	if in.Completed != nil {
		patch.Completed = in.Completed
	}
	if patch.Description == nil && patch.DueDate == nil && patch.Completed == nil {
		return nil, fmt.Errorf("%w: no fields to update", domain.ErrValidation)
	}
	if err := s.repo.Update(ctx, id, patch); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *TodoService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func validateDescription(description string) (string, error) {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return "", fmt.Errorf("%w: description is required", domain.ErrValidation)
	}
	return trimmed, nil
}

func parseDate(raw string) (time.Time, error) {
	d, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: due_date must be YYYY-MM-DD", domain.ErrValidation)
	}
	return d, nil
}

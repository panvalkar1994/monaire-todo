package service_test

import (
	"context"
	"testing"

	"todo/internal/domain"
	"todo/internal/service"
)

type mockRepo struct {
	items map[string]*domain.Todo
}

func newMockRepo() *mockRepo {
	return &mockRepo{items: map[string]*domain.Todo{}}
}

func (m *mockRepo) Create(_ context.Context, todo *domain.Todo) error {
	copied := *todo
	m.items[todo.ID] = &copied
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id string) (*domain.Todo, error) {
	t, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *t
	return &copied, nil
}

func (m *mockRepo) List(_ context.Context, includeCompleted bool) ([]*domain.Todo, error) {
	var out []*domain.Todo
	for _, t := range m.items {
		if !includeCompleted && t.Completed {
			continue
		}
		copied := *t
		out = append(out, &copied)
	}
	return out, nil
}

func (m *mockRepo) Replace(_ context.Context, todo *domain.Todo) error {
	existing, ok := m.items[todo.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if existing.Description == todo.Description && existing.DueDate.Equal(todo.DueDate) && existing.Completed == todo.Completed {
		return domain.ErrNoChanges
	}
	copied := *todo
	m.items[todo.ID] = &copied
	return nil
}

func (m *mockRepo) Update(_ context.Context, id string, patch domain.TodoPatch) error {
	t, ok := m.items[id]
	if !ok {
		return domain.ErrNotFound
	}
	if patch.Description != nil {
		t.Description = *patch.Description
	}
	if patch.DueDate != nil {
		t.DueDate = *patch.DueDate
	}
	if patch.Completed != nil {
		t.Completed = *patch.Completed
	}
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func TestCreateValidation(t *testing.T) {
	svc := service.NewTodoService(newMockRepo())
	_, err := svc.Create(context.Background(), service.CreateInput{Description: "  ", DueDate: "2026-06-15"})
	if err == nil {
		t.Fatal("expected validation error for empty description")
	}
	_, err = svc.Create(context.Background(), service.CreateInput{Description: "ok", DueDate: "bad-date"})
	if err == nil {
		t.Fatal("expected validation error for bad date")
	}
}

func TestCreateSuccess(t *testing.T) {
	svc := service.NewTodoService(newMockRepo())
	todo, err := svc.Create(context.Background(), service.CreateInput{Description: "buy milk", DueDate: "2026-06-15"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if todo.ID == "" || todo.Completed {
		t.Fatalf("unexpected todo: %+v", todo)
	}
}

func TestReplaceCreatesWhenMissing(t *testing.T) {
	svc := service.NewTodoService(newMockRepo())
	result, err := svc.Replace(context.Background(), "fixed-id", service.ReplaceInput{
		Description: "new task", DueDate: "2026-06-20", Completed: true,
	})
	if err != nil {
		t.Fatalf("replace create: %v", err)
	}
	if result.Todo.ID != "fixed-id" || !result.Todo.Completed || result.NoChanges {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestReplaceNoChanges(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewTodoService(repo)
	created, err := svc.Create(context.Background(), service.CreateInput{Description: "same", DueDate: "2026-06-15"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	result, err := svc.Replace(context.Background(), created.ID, service.ReplaceInput{
		Description: "same", DueDate: "2026-06-15", Completed: false,
	})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}
	if !result.NoChanges {
		t.Fatal("expected NoChanges=true")
	}
}

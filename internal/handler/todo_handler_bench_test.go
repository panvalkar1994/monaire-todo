package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"todo/internal/domain"
	"todo/internal/handler"
	"todo/internal/service"
)

type benchRepo struct{}

func (benchRepo) Create(_ context.Context, t *domain.Todo) error { return nil }
func (benchRepo) GetByID(_ context.Context, id string) (*domain.Todo, error) {
	return &domain.Todo{ID: id, Text: "x", DueDate: time.Now(), Completed: false}, nil
}
func (benchRepo) List(_ context.Context, _ bool) ([]*domain.Todo, error) {
	return []*domain.Todo{{ID: "1", Text: "a", DueDate: time.Now()}}, nil
}
func (benchRepo) Replace(_ context.Context, _ *domain.Todo) error { return nil }
func (benchRepo) Update(_ context.Context, _ string, _ domain.TodoPatch) error { return nil }
func (benchRepo) Delete(_ context.Context, _ string) error { return nil }

func BenchmarkListTodos(b *testing.B) {
	svc := service.NewTodoService(benchRepo{})
	r := handler.NewRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func BenchmarkCreateTodo(b *testing.B) {
	svc := service.NewTodoService(benchRepo{})
	r := handler.NewRouter(svc)
	body := []byte(`{"text":"bench","due_date":"2026-06-15"}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}
}

package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"todo/internal/domain"
	"todo/internal/service"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	svc *service.TodoService
}

func NewTodoHandler(svc *service.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

func (h *TodoHandler) Register(r gin.IRoutes) {
	r.POST("/todos", h.create)
	r.GET("/todos", h.list)
	r.GET("/todos/:id", h.get)
	r.PUT("/todos/:id", h.replace)
	r.PATCH("/todos/:id", h.patch)
	r.DELETE("/todos/:id", h.delete)
}

func (h *TodoHandler) create(c *gin.Context) {
	var req createTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	todo, err := h.svc.Create(c.Request.Context(), service.CreateInput{
		Description: req.Description,
		DueDate:   req.DueDate,
		Completed: req.Completed,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toResponse(todo))
}

func (h *TodoHandler) get(c *gin.Context) {
	todo, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(todo))
}

func (h *TodoHandler) list(c *gin.Context) {
	includeCompleted, err := parseIncludeCompleted(c.Query("include_completed"))
	if err != nil {
		writeError(c, http.StatusExpectationFailed, err.Error())
		return
	}
	todos, err := h.svc.List(c.Request.Context(), includeCompleted)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	if todos == nil {
		todos = []*domain.Todo{}
	}
	out := make([]todoResponse, 0, len(todos))
	for _, t := range todos {
		out = append(out, toResponse(t))
	}
	c.JSON(http.StatusOK, out)
}

func (h *TodoHandler) replace(c *gin.Context) {
	var req replaceTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	result, err := h.svc.Replace(c.Request.Context(), c.Param("id"), service.ReplaceInput{
		Description: req.Description,
		DueDate:   req.DueDate,
		Completed: req.Completed,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	if result.NoChanges {
		c.Header("X-No-Changes", "true")
	}
	c.JSON(http.StatusOK, toResponse(result.Todo))
}

func (h *TodoHandler) patch(c *gin.Context) {
	var req patchTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	todo, err := h.svc.Patch(c.Request.Context(), c.Param("id"), service.PatchInput{
		Description: req.Description,
		DueDate:   req.DueDate,
		Completed: req.Completed,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(todo))
}

func (h *TodoHandler) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func toResponse(t *domain.Todo) todoResponse {
	return todoResponse{
		ID:          t.ID,
		Description: t.Description,
		DueDate:   t.DueDate.Format("2006-01-02"),
		Completed: t.Completed,
	}
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(c, http.StatusNotFound, "todo not found")
	case errors.Is(err, domain.ErrValidation):
		writeError(c, http.StatusBadRequest, err.Error())
	default:
		slog.Error("request failed",
			"error", err,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		writeError(c, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(c *gin.Context, code int, msg string) {
	c.JSON(code, errorResponse{Error: msg})
}

const includeCompletedAllowedMsg = "include_completed allowed values: true|false|empty"

func parseIncludeCompleted(raw string) (bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, nil
	}
	switch strings.ToLower(raw) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, errors.New(includeCompletedAllowedMsg)
	}
}

package handler

import (
	"log/slog"
	"net/http"
	"time"

	"todo/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(svc *service.TodoService) *gin.Engine {
	r := gin.New()
	r.Use(slogAccessLog(), slogRecovery())

	h := NewTodoHandler(svc)
	api := r.Group("/api/v1")
	h.Register(api)

	return r
}

func slogAccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		if query != "" {
			path = path + "?" + query
		}
		slog.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
			"client_ip", c.ClientIP(),
		)
	}
}

func slogRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic recovered",
			"error", recovered,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

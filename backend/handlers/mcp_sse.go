package handlers

import (
	"mcp-adapter/backend/adapter"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServeSSE(c *gin.Context) {
	path := c.Param("path")
	impl := adapter.GetServerImpl(path)
	if impl == nil {
		c.String(http.StatusNotFound, "sse path not found")
		return
	}
	// 复用底层 http.Handler
	impl.ServeHTTP(c.Writer, c.Request)
}

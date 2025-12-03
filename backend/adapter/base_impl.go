package adapter

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type RequestHandle interface {
	// DoRequest 执行MCP实际请求
	DoRequest(ctx context.Context, req mcp.CallToolRequest, parameters Parameters, meta RequestMeta) ([]byte, error)
	// Compatible 检查请求是否兼容当前处理器
	Compatible(meta RequestMeta) bool
}

package main

import (
	"context"
	"errors"
	"log"
	"mcp-adapter/backend/adapter"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/routes"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	// 初始化数据库
	database.InitDatabase("mcp-adapter.db")

	// 确保数据库连接在程序退出时关闭
	defer func() {
		if sqlDB, err := database.GetDB().DB(); err == nil {
			if closeErr := sqlDB.Close(); closeErr != nil {
				log.Printf("Error closing database: %v", closeErr)
			} else {
				log.Println("Database connection closed")
			}
		}
	}()

	// 初始化默认数据（首次启动时创建 MCP-Adapter 应用及其接口）
	database.InitDefaultData()

	// 初始化 MCP 服务器管理器
	adapter.InitServer()

	// 设置路由
	router := routes.SetupRoutes()

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 在 goroutine 中启动服务器
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT (Ctrl+C) 和 SIGTERM 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	// 先关闭 adapter，停止事件处理
	adapter.Shutdown()

	// 创建 5 秒超时的 context 用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
	}
}

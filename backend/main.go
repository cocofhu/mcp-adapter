package main

import (
	"log"
	"mcp-adapter/backend/adapter"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/routes"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {
	// 初始化数据库
	database.InitDatabase("mcp-adapter.db")

	// 设置路由
	router := routes.SetupRoutes()
	// 初始化
	adapter.InitServer()
	// 启动服务器
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

package main

import (
	"log"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/routes"
	"net/http"
)

func main() {
	// 初始化数据库
	database.InitDatabase()

	// 设置路由
	router := routes.SetupRoutes()

	// 启动服务器
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

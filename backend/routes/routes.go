package routes

import (
	"log"
	"mcp-adapter/backend/handlers"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// SetupRoutes 设置路由
func SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	// 记录日志
	r.Use(loggingMiddleware)
	// 启用CORS
	r.Use(corsMiddleware)
	// API路由组
	api := r.PathPrefix("/api").Subrouter()

	// 应用相关路由
	api.HandleFunc("/applications", handlers.CreateApplication).Methods("POST")
	api.HandleFunc("/applications", handlers.GetApplications).Methods("GET")
	api.HandleFunc("/applications/{id}", handlers.GetApplication).Methods("GET")
	api.HandleFunc("/applications/{id}", handlers.UpdateApplication).Methods("PUT")
	api.HandleFunc("/applications/{id}", handlers.DeleteApplication).Methods("DELETE")

	// 接口相关路由
	api.HandleFunc("/interfaces", handlers.CreateInterface).Methods("POST")
	api.HandleFunc("/interfaces", handlers.GetInterfaces).Methods("GET")
	api.HandleFunc("/interfaces/{id}", handlers.GetInterface).Methods("GET")
	api.HandleFunc("/interfaces/{id}", handlers.UpdateInterface).Methods("PUT")
	api.HandleFunc("/interfaces/{id}", handlers.DeleteInterface).Methods("DELETE")

	// 静态文件服务
	// 静态目录
	staticFileDirectory := http.Dir("./web/static")
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/static/").Handler(staticFileHandler)

	// 主页
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/index.html")
	})

	// MCP-SSE服务
	r.HandleFunc("/sse/{path}", handlers.ServeSSE)
	r.HandleFunc("/message/{path}", handlers.ServeSSE)

	return r
}

// corsMiddleware CORS中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
		log.Printf("%s %s, Request processed in %v", r.Method, r.RequestURI, time.Since(start))
	})
}

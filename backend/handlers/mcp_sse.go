package handlers

import (
	"mcp-adapter/backend/adapter"
	"net/http"

	"github.com/gorilla/mux"
)

func ServeSSE(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	impl := adapter.GetServerImpl(path)
	if impl == nil {
		http.Error(w, "sse path not found", http.StatusNotFound)
		return
	}
	impl.ServeHTTP(w, r)
}

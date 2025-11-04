package handlers

import (
	"encoding/json"
	"mcp-adapter/backend/service"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateInterface 创建接口
func CreateInterface(w http.ResponseWriter, r *http.Request) {
	var req service.CreateInterfaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	resp, err := service.CreateInterface(req)
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Application not found", http.StatusBadRequest)
		case service.ErrIfaceNameExists:
			http.Error(w, "Interface name already exists", http.StatusBadRequest)
		case service.ErrInvalidOptions:
			http.Error(w, "Options validation failed", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to create interface", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Interface)
}

// GetInterfaces 获取所有接口
func GetInterfaces(w http.ResponseWriter, r *http.Request) {
	var req service.ListInterfacesRequest
	if appIDStr := r.URL.Query().Get("app_id"); appIDStr != "" {
		if id, err := strconv.ParseInt(appIDStr, 10, 64); err == nil {
			req.AppID = &id
		}
	}
	resp, err := service.ListInterfaces(req)
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to fetch interfaces", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Interfaces)
}

// GetInterface 获取单个接口
func GetInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		return
	}

	resp, err := service.GetInterface(service.GetInterfaceRequest{ID: id})
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Interface not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to fetch interface", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Interface)
}

// UpdateInterface 更新接口（部分字段）
func UpdateInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		return
	}

	var body service.UpdateInterfaceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	body.ID = id

	resp, err := service.UpdateInterface(body)
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Interface or Application not found", http.StatusNotFound)
		case service.ErrIfaceNameExists:
			http.Error(w, "Interface name already exists", http.StatusBadRequest)
		case service.ErrInvalidOptions:
			http.Error(w, "Options validation failed", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update interface", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Interface)
}

// DeleteInterface 删除接口
func DeleteInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		return
	}

	_, err = service.DeleteInterface(service.DeleteInterfaceRequest{ID: id})
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Interface not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to delete interface", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

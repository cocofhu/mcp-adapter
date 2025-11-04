package handlers

import (
	"encoding/json"
	"mcp-adapter/backend/service"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateApplication 创建应用
func CreateApplication(w http.ResponseWriter, r *http.Request) {
	var req service.CreateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	resp, err := service.CreateApplication(req)
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrAppNameExists:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Failed to create application", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Application)
}

// GetApplications 获取所有应用
func GetApplications(w http.ResponseWriter, r *http.Request) {
	resp, err := service.ListApplications(service.ListApplicationsRequest{})
	if err != nil {
		http.Error(w, "Failed to fetch applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Applications)
}

// GetApplication 获取单个应用
func GetApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}

	resp, err := service.GetApplication(service.GetApplicationRequest{ID: id})
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid application ID", http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Application not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to fetch application", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Application)
}

// UpdateApplication 更新应用（部分字段）
func UpdateApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}

	var body service.UpdateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	body.ID = id

	resp, err := service.UpdateApplication(body)
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Application not found", http.StatusNotFound)
		case service.ErrAppNameExists:
			http.Error(w, "Application name already exists", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update application", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp.Application)
}

// DeleteApplication 删除应用
func DeleteApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}

	_, err = service.DeleteApplication(service.DeleteApplicationRequest{ID: id})
	if err != nil {
		switch err {
		case service.ErrValidation:
			http.Error(w, "Invalid application ID", http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Application not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to delete application", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

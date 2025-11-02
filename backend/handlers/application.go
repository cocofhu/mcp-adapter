package handlers

import (
	"encoding/json"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateApplication 创建应用
func CreateApplication(w http.ResponseWriter, r *http.Request) {
	var app models.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	// 判断应用名字是否重复（全局唯一）
	var count int64
	db.Model(&models.Application{}).Where("name = ?", app.Name).Count(&count)
	if count > 0 {
		http.Error(w, "Application name already exists", http.StatusBadRequest)
		return
	}

	if err := db.Create(&app).Error; err != nil {
		http.Error(w, "Failed to create application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(app)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// GetApplications 获取所有应用
func GetApplications(w http.ResponseWriter, r *http.Request) {
	var apps []models.Application
	db := database.GetDB()

	if err := db.Find(&apps).Error; err != nil {
		http.Error(w, "Failed to fetch applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(apps)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// GetApplication 获取单个应用
func GetApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}

	var app models.Application
	db := database.GetDB()

	if err := db.First(&app, id).Error; err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(app)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// UpdateApplication 更新应用
func UpdateApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}

	var app models.Application
	db := database.GetDB()

	if err := db.First(&app, id).Error; err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	var updateData models.Application
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 保持原有ID
	updateData.ID = app.ID

	// 如果更新了名称，校验是否重复（全局唯一）
	newName := updateData.Name
	if newName == "" {
		newName = app.Name
	}
	var cnt int64
	db.Model(&models.Application{}).Where("name = ? AND id <> ?", newName, app.ID).Count(&cnt)
	if cnt > 0 {
		http.Error(w, "Application name already exists", http.StatusBadRequest)
		return
	}

	if err := db.Save(&updateData).Error; err != nil {
		http.Error(w, "Failed to update application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updateData)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// DeleteApplication 删除应用
func DeleteApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()

	// 检查应用是否存在
	var app models.Application
	if err := db.First(&app, id).Error; err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	// 删除应用及其关联的接口
	if err := db.Delete(&app).Error; err != nil {
		http.Error(w, "Failed to delete application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

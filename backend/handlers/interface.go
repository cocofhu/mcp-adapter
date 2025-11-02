package handlers

import (
	"encoding/json"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateInterface 创建接口
func CreateInterface(w http.ResponseWriter, r *http.Request) {
	var iface models.Interface
	if err := json.NewDecoder(r.Body).Decode(&iface); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	db := database.GetDB()

	// 验证应用是否存在
	var app models.Application
	if err := db.First(&app, iface.AppID).Error; err != nil {
		http.Error(w, "Application not found", http.StatusBadRequest)
		return
	}

	// 判断接口名字是否重复（同一应用下唯一）
	var count int64
	db.Model(&models.Interface{}).Where("app_id = ? AND name = ?", iface.AppID, iface.Name).Count(&count)
	if count > 0 {
		http.Error(w, "Interface name already exists", http.StatusBadRequest)
		return
	}

	options, err := iface.GetToolOptions()
	if err != nil {
		http.Error(w, "Failed to get options", http.StatusBadRequest)
	}

	err = options.Validate()
	if err != nil {
		http.Error(w, "Options validation failed", http.StatusBadRequest)
	}

	if err := db.Create(&iface).Error; err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(iface)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// GetInterfaces 获取所有接口
func GetInterfaces(w http.ResponseWriter, r *http.Request) {
	var interfaces []models.Interface
	db := database.GetDB()

	// 支持按应用ID过滤
	appID := r.URL.Query().Get("app_id")
	query := db
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if err := query.Find(&interfaces).Error; err != nil {
		http.Error(w, "Failed to fetch interfaces", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(interfaces)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// GetInterface 获取单个接口
func GetInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		return
	}

	var iface models.Interface
	db := database.GetDB()

	if err := db.First(&iface, id).Error; err != nil {
		http.Error(w, "Interface not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(iface)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// UpdateInterface 更新接口
func UpdateInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		return
	}

	var iface models.Interface
	db := database.GetDB()

	if err := db.First(&iface, id).Error; err != nil {
		http.Error(w, "Interface not found", http.StatusNotFound)
		return
	}

	var updateData models.Interface
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 如果更新了应用ID，验证应用是否存在
	if updateData.AppID != 0 && updateData.AppID != iface.AppID {
		var app models.Application
		if err := db.First(&app, updateData.AppID).Error; err != nil {
			http.Error(w, "Application not found", http.StatusBadRequest)
			return
		}
	}

	options, err := iface.GetToolOptions()
	if err != nil {
		http.Error(w, "Failed to get options", http.StatusBadRequest)
	}

	err = options.Validate()
	if err != nil {
		http.Error(w, "Options validation failed", http.StatusBadRequest)
	}

	// 保持原有ID
	updateData.ID = iface.ID

	// 如果更新了名称或AppID，校验同一应用下名称唯一
	newAppID := updateData.AppID
	if newAppID == 0 {
		newAppID = iface.AppID
	}
	newName := updateData.Name
	if newName == "" {
		newName = iface.Name
	}
	var cnt int64
	db.Model(&models.Interface{}).Where("app_id = ? AND name = ? AND id <> ?", newAppID, newName, iface.ID).Count(&cnt)
	if cnt > 0 {
		http.Error(w, "Interface name already exists", http.StatusBadRequest)
		return
	}

	if err := db.Save(&updateData).Error; err != nil {
		http.Error(w, "Failed to update interface", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updateData)
	if err != nil {
		http.Error(w, "Failed to create interface", http.StatusInternalServerError)
	}
}

// DeleteInterface 删除接口
func DeleteInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid interface ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()

	// 检查接口是否存在
	var iface models.Interface
	if err := db.First(&iface, id).Error; err != nil {
		http.Error(w, "Interface not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&iface).Error; err != nil {
		http.Error(w, "Failed to delete interface", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

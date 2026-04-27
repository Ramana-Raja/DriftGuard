package models

import (
	"DriftGuard/backend/database"
	"DriftGuard/backend/storage"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetProjectModel(c *gin.Context) {
	userID := c.MustGet("currentUser").(uint)

	idParam := c.Param("id")
	projectID, _ := strconv.Atoi(idParam)

	var project database.Project

	err := database.DB.
		Joins("JOIN user_profiles ON user_profiles.id = projects.user_profile_id").
		Where("projects.id = ? AND user_profiles.user_auth_id = ?", projectID, userID).
		Preload("ProjectModel.Versions").
		First(&project).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var activeVersion database.ModelVersion

	for _, v := range project.ProjectModel.Versions {
		if v.IsActive {
			activeVersion = v
			break
		}
	}

	c.JSON(200, gin.H{
		"active_model": activeVersion,
		"history":      project.ProjectModel.Versions,
	})
}

func UploadModel(c *gin.Context) {
	userID := c.MustGet("currentUser").(uint)

	projectIDParam := c.Param("id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid project id"})
		return
	}

	fileHeader, err := c.FormFile("model")
	if err != nil {
		c.JSON(400, gin.H{"error": "file required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(fileHeader.Filename, ".pkl") {
		c.JSON(400, gin.H{"error": "only .pkl files allowed"})
		return
	}
	var project database.Project
	err = database.DB.
		Joins("JOIN user_profiles ON user_profiles.id = projects.user_profile_id").
		Where("projects.id = ? AND user_profiles.user_auth_id = ?", projectID, userID).
		First(&project).Error

	if err != nil {
		c.JSON(403, gin.H{"error": "unauthorized"})
		return
	}

	var model database.Model
	err = database.DB.Where("project_id = ?", project.ID).First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		model = database.Model{
			ProjectID: project.ID,
			Name:      "default-model",
		}

		if err := database.DB.Create(&model).Error; err != nil {
			c.JSON(500, gin.H{"error": "failed to create model"})
			return
		}

	} else if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	var last database.ModelVersion

	err = database.DB.
		Where("model_id = ?", model.ID).
		Order("version desc").
		First(&last).Error

	var newVersion int

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newVersion = 1
	} else if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	} else {
		newVersion = last.Version + 1
	}

	objectName := fmt.Sprintf("project-%d/model-%d/v%d.pkl", project.ID, model.ID, newVersion)

	path, err := storage.UploadFile("models", objectName, file, fileHeader.Size)
	if err != nil {
		c.JSON(500, gin.H{"error": "upload failed"})
		return
	}

	if err := database.DB.
		Model(&database.ModelVersion{}).
		Where("model_id = ?", model.ID).
		Update("is_active", false).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to update versions"})
		return
	}
	version := database.ModelVersion{
		ModelID:  model.ID,
		Version:  newVersion,
		FilePath: path,
		IsActive: true,
	}

	if err := database.DB.Create(&version).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to save version"})
		return
	}

	c.JSON(200, gin.H{
		"message": "model uploaded successfully",
		"version": newVersion,
	})
}

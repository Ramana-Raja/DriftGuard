package models

import (
	"DriftGuard/backend/database"

	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProejcts(c *gin.Context) {
	userID := c.MustGet("currentUser").(uint)

	var projects []database.Project

	err := database.DB.
		Joins("JOIN user_profiles ON user_profiles.id = projects.user_profile_id").
		Where("user_profiles.user_auth_id = ?", userID).
		Find(&projects).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while getting projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

type CreateProjectInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func CreateProjects(c *gin.Context) {
	userID := c.MustGet("currentUser").(uint)
	var project CreateProjectInput

	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var userprofie database.UserProfile

	err := database.DB.Select("id").Where("user_auth_id = ?", userID).First(&userprofie).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile not found"})
		return
	}

	input := database.Project{
		UserProfileID: userprofie.ID,
		ProjectName:   project.Name,
		Description:   project.Description,
	}
	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, input)
}

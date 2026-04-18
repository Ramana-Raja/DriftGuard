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

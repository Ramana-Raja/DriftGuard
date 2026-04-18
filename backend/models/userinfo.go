package models

import (
	"DriftGuard/backend/database"

	"github.com/gin-gonic/gin"
	"net/http"
)

func Getemail(c *gin.Context) {
	userID := c.MustGet("currentUser").(uint)

	var email string

	err := database.DB.Model(&database.UserAuth{}).Select("email").Where("id = ?", userID).Scan(&email).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no email found in database for this jwt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"email": email})
}

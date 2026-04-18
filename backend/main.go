package main

import (
	"DriftGuard/backend/controllers"
	"DriftGuard/backend/database"
	"DriftGuard/backend/middleware"
	"DriftGuard/backend/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "myuser"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "mypassword"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "myapp"
	}

	database.NewDB(host, port, user, password, dbname)

	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"Authorization", "Content-Type"},
	}))
	r.POST("/api/register", controllers.Register)
	r.POST("/api/login", controllers.Login)

	protected := r.Group("/api")

	protected.Use(middleware.JwtHandler())
	{
		protected.GET("/me", models.Getemail)
		protected.GET("/projects", models.GetProejcts)
	}
	r.Run(":8080")
}

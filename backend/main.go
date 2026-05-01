package main

import (
	"DriftGuard/backend/controllers"
	"DriftGuard/backend/database"
	"DriftGuard/backend/middleware"
	"DriftGuard/backend/models"
	"DriftGuard/backend/storage"
	"DriftGuard/backend/tasks"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	//database.ResetDB()
	storage.InitMinIO()
	tasks.RedisInt()

	go tasks.StartDriftCheck()
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
		protected.POST("/projects", models.CreateProjects)

		protected.GET("/projects/:id/model", models.GetProjectModel)
		protected.POST("/projects/:id/model", models.UploadModel)
	}
	r.Run(":8080")
}

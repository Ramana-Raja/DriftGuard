package main

import (
	"DriftGuard/backend/controllers"
	"DriftGuard/backend/database"
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

	db, err := database.NewDB(host, port, user, password, dbname)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer db.Close()

	r := gin.Default()
	r.Use(cors.Default())
	r.POST("/api/register", controllers.Register(db))

	r.Run(":8080")
}

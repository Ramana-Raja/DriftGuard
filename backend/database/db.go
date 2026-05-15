package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ResetDB() {
	err := DB.Migrator().DropTable(
		&ModelVersion{},
		&Model{},
		&Project{},
		&UserProfile{},
		&UserAuth{},
	)
	if err != nil {
		return
	}

	err = DB.AutoMigrate(
		&UserAuth{},
		&UserProfile{},
		&Project{},
		&Model{},
		&ModelVersion{},
	)
	if err != nil {
		return
	}
}
func NewDB(host, port, user, password, dbname string) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}

	err = conn.AutoMigrate(
		&UserAuth{},
		&UserProfile{},
		&Project{},
		&Model{},
		&ModelVersion{},
	)

	if err != nil {
		return
	}
	log.Println("Connection successful to database")
	DB = conn

	sqlDB, err := conn.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetMaxOpenConns(40)
		sqlDB.SetConnMaxLifetime(time.Minute * 5)
		log.Println("Database connection pooling optimized for Azure PostgreSQL limits.")
	} else {
		log.Printf("Warning: Failed to retrieve generic SQL driver block for pooling settings: %v", err)
	}
}

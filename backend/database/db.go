package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type DB struct {
	conn *gorm.DB
}

func NewDB(host, port, user, password, dbname string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Println("Connection successful to database")
	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

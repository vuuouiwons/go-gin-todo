package data

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a SQLite connection and automatically migrates all schema tables.
func Connect(dbPath string) (*gorm.DB, error) {
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Printf("Failed to open SQLite database: %v\n", err)
		return nil, err
	}

	if err := database.AutoMigrate(&UserEntity{}, &TodolistEntity{}, &TodoEntity{}); err != nil {
		log.Printf("Failed to run database auto-migrations: %v\n", err)
		return nil, err
	}

	log.Println("SQLite database connected and migrations applied successfully")
	return database, nil
}

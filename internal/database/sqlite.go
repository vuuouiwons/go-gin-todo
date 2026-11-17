package database

import (
	"log"

	"github.com/glebarez/sqlite"
	"go-gin-todo/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSQLiteDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("Failed to open SQLite database: %v\n", err)
		return nil, err
	}

	// AutoMigrate the domain entities
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Todolist{},
		&domain.Todo{},
	)
	if err != nil {
		log.Printf("Failed to run database auto-migrations: %v\n", err)
		return nil, err
	}

	log.Println("SQLite database connected and migrations applied successfully")
	return db, nil
}

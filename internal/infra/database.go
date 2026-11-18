package infra

import (
	"context"
	"log"

	"github.com/glebarez/sqlite"
	"go-gin-todo/internal/app"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type txKey struct{}

// NewSQLiteDatabase opens a SQLite connection and auto-migrates application models.
func NewSQLiteDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("Failed to open SQLite database: %v\n", err)
		return nil, err
	}

	if err := db.AutoMigrate(&app.User{}, &app.Todolist{}, &app.Todo{}); err != nil {
		log.Printf("Failed to run database auto-migrations: %v\n", err)
		return nil, err
	}

	log.Println("SQLite database connected and migrations applied successfully")
	return db, nil
}

// dbFromContext returns the active transaction DB if present in ctx, otherwise defaultDB.
func dbFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}

// Database implements app.Transactor and app.Queryer using GORM.
type Database struct {
	db *gorm.DB
}

// NewDatabase wraps a *gorm.DB instance to provide Transactor and Queryer capabilities.
func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

// DB implements app.Queryer for direct queries in Application Queries.
func (d *Database) DB(ctx context.Context) *gorm.DB {
	return dbFromContext(ctx, d.db)
}

// ExecInTransaction implements app.Transactor with Propagation.Required semantics:
// reuses an existing transaction in ctx if present, otherwise starts a new one.
func (d *Database) ExecInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return fn(ctx)
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

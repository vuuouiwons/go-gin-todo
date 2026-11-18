package infra

import (
	"context"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"go-gin-todo/internal/app"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type txKey struct{}
type txIDKey struct{}

// txSQLiteLogger logs SQL queries prefixed with the active transaction UUID from ctx.
type txSQLiteLogger struct{}

func (l *txSQLiteLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *txSQLiteLogger) Info(_ context.Context, msg string, data ...interface{}) {
	log.Printf(msg, data...)
}

func (l *txSQLiteLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	log.Printf(msg, data...)
}

func (l *txSQLiteLogger) Error(_ context.Context, msg string, data ...interface{}) {
	log.Printf(msg, data...)
}

func (l *txSQLiteLogger) Trace(ctx context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sqlStr, _ := fc()
	txID, _ := ctx.Value(txIDKey{}).(string)
	if txID == "" {
		txID = "default"
	}
	log.Printf("Executing (%s): %s;", txID, sqlStr)
}

// NewSQLiteDatabase opens a SQLite connection and auto-migrates application models.
func NewSQLiteDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 &txSQLiteLogger{},
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

	txID := uuid.NewString()
	log.Printf("Executing (%s): BEGIN DEFERRED TRANSACTION;", txID)

	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	txCtx = context.WithValue(txCtx, txIDKey{}, txID)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Executing (%s): ROLLBACK;", txID)
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		tx.Rollback()
		log.Printf("Executing (%s): ROLLBACK;", txID)
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Printf("Executing (%s): ROLLBACK;", txID)
		return err
	}

	log.Printf("Executing (%s): COMMIT;", txID)
	return nil
}

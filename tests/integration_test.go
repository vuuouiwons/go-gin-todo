package tests

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"go-gin-todo/src/db"
	"go-gin-todo/src/db/repository"
	"go-gin-todo/src/services"
	"go-gin-todo/src/utils"
	"gorm.io/gorm"
)

func setupIsolatedIntegrationDB(t *testing.T) (*gorm.DB, *services.AuthService, *services.TodolistService, *services.TodoService) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "integration.db")
	database, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to create isolated test DB: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := database.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	userRepo := repository.NewUserRepository(database)
	todolistRepo := repository.NewTodolistRepository(database)
	todoRepo := repository.NewTodoRepository(database)

	authService := services.NewAuthService(userRepo, "jwt-secret", 2)
	todolistService := services.NewTodolistService(database, todolistRepo)
	todoService := services.NewTodoService(database, todoRepo)

	return database, authService, todolistService, todoService
}

func TestIntegration_UserRepositoryConstraints(t *testing.T) {
	ctx := context.Background()
	database, authService, _, _ := setupIsolatedIntegrationDB(t)

	u, err := authService.Register(ctx, "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("register user failed: %v", err)
	}
	if u.ID == 0 {
		t.Fatalf("expected auto-incremented ID")
	}

	// Duplicate email violates unique index
	_, err = authService.Register(ctx, "alice@example.com", "otherpassword")
	if !errors.Is(err, utils.ErrConflict) {
		t.Fatalf("expected utils.ErrConflict on duplicate email, got: %v", err)
	}

	// Verify DB record directly
	var found db.User
	if err := database.Where("email = ?", "alice@example.com").First(&found).Error; err != nil || found.ID != u.ID {
		t.Fatalf("database query failed: got=%+v err=%v", found, err)
	}
}

func TestIntegration_TodolistQueriesAndTransactionRollback(t *testing.T) {
	ctx := context.Background()
	_, _, todolistService, _ := setupIsolatedIntegrationDB(t)

	// Seed data for user 1 and user 2
	_, err := todolistService.CreateBulk(ctx, 1, []string{
		"Alpha Project",
		"Beta Groceries",
		"Gamma Project",
	})
	if err != nil {
		t.Fatalf("seed CreateBulk failed: %v", err)
	}

	_, err = todolistService.CreateBulk(ctx, 2, []string{
		"User2 Project",
	})
	if err != nil {
		t.Fatalf("seed user 2 failed: %v", err)
	}

	// Query pagination & ordering (id DESC)
	page1, meta1, err := todolistService.GetBulk(ctx, 1, "", 1, 2)
	if err != nil {
		t.Fatalf("GetBulk page 1 failed: %v", err)
	}
	if len(page1) != 2 || meta1.TotalItems != 3 || meta1.TotalPages != 2 || !meta1.HasNext || meta1.HasPrev {
		t.Fatalf("unexpected page1 result/meta: len=%d meta=%+v", len(page1), meta1)
	}
	if page1[0].Title != "Gamma Project" {
		t.Fatalf("expected newest item first (id DESC), got %q", page1[0].Title)
	}

	// Query search filter
	filtered, metaFiltered, err := todolistService.GetBulk(ctx, 1, "Project", 1, 10)
	if err != nil {
		t.Fatalf("GetBulk search failed: %v", err)
	}
	if len(filtered) != 2 || metaFiltered.TotalItems != 2 {
		t.Fatalf("expected 2 matching Project items for user 1, got %d", len(filtered))
	}

	// Real DB transaction rollback test via TodolistService.UpdateBulk
	_, err = todolistService.UpdateBulk(ctx, 1, []db.Todolist{
		{ID: filtered[0].ID, Title: "Mutated Then Rolled Back"},
		{ID: 99999, Title: "Nonexistent"},
	})
	if !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	persisted, err := todolistService.GetByID(ctx, filtered[0].ID, 1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if persisted.Title == "Mutated Then Rolled Back" {
		t.Fatalf("expected DB transaction rollback to revert title change, but change persisted")
	}
}

func TestIntegration_TodoQueriesAndFilters(t *testing.T) {
	ctx := context.Background()
	_, _, _, todoService := setupIsolatedIntegrationDB(t)

	_, err := todoService.CreateBulk(ctx, 1, []db.Todo{
		{Message: "Buy organic milk", Check: false},
		{Message: "Write integration tests", Check: true},
		{Message: "Review PRs", Check: true},
	})
	if err != nil {
		t.Fatalf("seed todos failed: %v", err)
	}

	_, err = todoService.CreateBulk(ctx, 2, []db.Todo{
		{Message: "Buy milk for user 2", Check: false},
	})
	if err != nil {
		t.Fatalf("seed user 2 todos failed: %v", err)
	}

	// Filter check=true
	onlyDone := true
	doneItems, doneMeta, err := todoService.GetBulk(ctx, 1, "", &onlyDone, 1, 10)
	if err != nil {
		t.Fatalf("GetBulk check=true failed: %v", err)
	}
	if len(doneItems) != 2 || doneMeta.TotalItems != 2 {
		t.Fatalf("expected 2 completed todos for user 1, got %d", len(doneItems))
	}

	// Filter check=false + search="milk"
	onlyOpen := false
	openMilk, milkMeta, err := todoService.GetBulk(ctx, 1, "milk", &onlyOpen, 1, 10)
	if err != nil {
		t.Fatalf("GetBulk check=false search=milk failed: %v", err)
	}
	if len(openMilk) != 1 || milkMeta.TotalItems != 1 || openMilk[0].Message != "Buy organic milk" {
		t.Fatalf("unexpected filtered todos: %+v", openMilk)
	}

	// Verify delete bulk
	deleted, err := todoService.DeleteBulk(ctx, []uint{doneItems[0].ID, doneItems[1].ID}, 1)
	if err != nil || deleted != 2 {
		t.Fatalf("expected 2 deleted, got %d err=%v", deleted, err)
	}
}

package tests

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"go-gin-todo/src/business"
	"go-gin-todo/src/data"
	"go-gin-todo/src/shared"
	"gorm.io/gorm"
)

func setupRepoTestDB(t *testing.T) (*gorm.DB, *data.UserRepository, *data.TodolistRepository, *data.TodoRepository) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "repo_test.db")
	database, err := data.Connect(dbPath)
	if err != nil {
		t.Fatalf("Failed to connect to test db: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, err := database.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return database, data.NewUserRepository(database), data.NewTodolistRepository(database), data.NewTodoRepository(database)
}

func TestUserRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	_, userRepo, _, _ := setupRepoTestDB(t)

	user := &business.User{Email: "bob@example.com", Password: "hashedpassword"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create user failed: %v", err)
	}
	if user.ID == 0 {
		t.Fatalf("Expected auto-generated ID, got 0")
	}

	found, err := userRepo.GetByEmail(ctx, "bob@example.com")
	if err != nil || found.ID != user.ID {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	byID, err := userRepo.GetByID(ctx, user.ID)
	if err != nil || byID.Email != "bob@example.com" {
		t.Fatalf("GetByID failed: %v", err)
	}

	_, err = userRepo.GetByEmail(ctx, "nonexistent@example.com")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound for missing email, got %v", err)
	}
}

func TestTodolistRepository_CRUDAndWithTx(t *testing.T) {
	ctx := context.Background()
	database, _, listRepo, _ := setupRepoTestDB(t)

	item := &business.Todolist{UserID: 1, Title: "List A"}
	if err := listRepo.Create(ctx, item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	bulk := []business.Todolist{
		{UserID: 1, Title: "List B"},
		{UserID: 1, Title: "List C"},
	}
	if err := listRepo.CreateBulk(ctx, &bulk); err != nil {
		t.Fatalf("CreateBulk failed: %v", err)
	}

	items, p, err := listRepo.GetBulk(ctx, 1, "", 1, 10)
	if err != nil || len(items) != 3 || p.TotalItems != 3 {
		t.Fatalf("GetBulk failed: %v", err)
	}

	byIDs, err := listRepo.FindByIDs(ctx, 1, []uint{item.ID, bulk[0].ID})
	if err != nil || len(byIDs) != 2 {
		t.Fatalf("FindByIDs failed: %v", err)
	}

	// Test WithTx rollback
	_ = database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := listRepo.WithTx(tx)
		item.Title = "Mutated Inside TX"
		_ = txRepo.Save(ctx, item)
		return errors.New("force rollback")
	})

	persisted, err := listRepo.GetByID(ctx, item.ID, 1)
	if err != nil || persisted.Title != "List A" {
		t.Fatalf("Expected rollback to preserve 'List A', got %q", persisted.Title)
	}

	deleted, err := listRepo.DeleteBulk(ctx, []uint{bulk[0].ID, bulk[1].ID}, 1)
	if err != nil || deleted != 2 {
		t.Fatalf("DeleteBulk failed: count=%d err=%v", deleted, err)
	}
}

func TestTodoRepository_CRUDAndWithTx(t *testing.T) {
	ctx := context.Background()
	database, _, _, todoRepo := setupRepoTestDB(t)

	todo := &business.Todo{UserID: 1, Message: "Task 1", Check: false}
	if err := todoRepo.Create(ctx, todo); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	bulk := []business.Todo{
		{UserID: 1, Message: "Task 2", Check: true},
		{UserID: 1, Message: "Task 3", Check: false},
	}
	if err := todoRepo.CreateBulk(ctx, &bulk); err != nil {
		t.Fatalf("CreateBulk failed: %v", err)
	}

	items, p, err := todoRepo.GetBulk(ctx, 1, "", nil, 1, 10)
	if err != nil || len(items) != 3 || p.TotalItems != 3 {
		t.Fatalf("GetBulk failed: %v", err)
	}

	byIDs, err := todoRepo.FindByIDs(ctx, 1, []uint{todo.ID, bulk[0].ID})
	if err != nil || len(byIDs) != 2 {
		t.Fatalf("FindByIDs failed: %v", err)
	}

	// Test WithTx rollback
	_ = database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := todoRepo.WithTx(tx)
		todo.Message = "Task 1 Mutated"
		_ = txRepo.Save(ctx, todo)
		return errors.New("force rollback")
	})

	persisted, err := todoRepo.GetByID(ctx, todo.ID, 1)
	if err != nil || persisted.Message != "Task 1" {
		t.Fatalf("Expected rollback to preserve 'Task 1', got %q", persisted.Message)
	}

	deleted, err := todoRepo.DeleteByID(ctx, todo.ID, 1)
	if err != nil || deleted != 1 {
		t.Fatalf("DeleteByID failed: %v", err)
	}
}

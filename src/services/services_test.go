package services

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"go-gin-todo/src/db"
	"go-gin-todo/src/db/repository"
	"go-gin-todo/src/utils"
)

func setupIsolatedServiceDB(t *testing.T) (*AuthService, *TodolistService, *TodoService) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test_service.db")
	database, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test db: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, err := database.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	userRepo := repository.NewUserRepository(database)
	todolistRepo := repository.NewTodolistRepository(database)
	todoRepo := repository.NewTodoRepository(database)

	authService := NewAuthService(userRepo, "test-secret-12345", 2)
	todolistService := NewTodolistService(database, todolistRepo)
	todoService := NewTodoService(database, todoRepo)

	return authService, todolistService, todoService
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	authService, _, _ := setupIsolatedServiceDB(t)

	// Valid registration
	u, err := authService.Register(ctx, "test@example.com", "secret123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if u.ID == 0 || u.Email != "test@example.com" {
		t.Fatalf("Unexpected user: %+v", u)
	}

	// Short password
	_, err = authService.Register(ctx, "test2@example.com", "123")
	if !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("Expected ErrBadRequest for short password, got: %v", err)
	}

	// Duplicate registration
	_, err = authService.Register(ctx, "test@example.com", "secret123")
	if !errors.Is(err, utils.ErrConflict) {
		t.Fatalf("Expected ErrConflict for duplicate email, got: %v", err)
	}

	// Successful login
	loggedInUser, token, err := authService.Login(ctx, "test@example.com", "secret123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loggedInUser.ID != u.ID || token == "" {
		t.Fatalf("Unexpected login result: user=%+v token=%q", loggedInUser, token)
	}

	// Invalid password login
	_, _, err = authService.Login(ctx, "test@example.com", "wrongpass")
	if !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatalf("Expected ErrUnauthorized for invalid password, got: %v", err)
	}

	// Non-existent user login
	_, _, err = authService.Login(ctx, "nonexistent@example.com", "secret123")
	if !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatalf("Expected ErrUnauthorized for missing user, got: %v", err)
	}
}

func TestTodolistService_TransactionRollback(t *testing.T) {
	ctx := context.Background()
	_, todolistService, _ := setupIsolatedServiceDB(t)

	// Create initial lists for user 1 and user 2
	list1, err := todolistService.CreateOne(ctx, 1, "List 1 Original")
	if err != nil {
		t.Fatalf("CreateOne failed: %v", err)
	}
	list2, err := todolistService.CreateOne(ctx, 2, "List 2 Belonging to User 2")
	if err != nil {
		t.Fatalf("CreateOne failed: %v", err)
	}

	// Attempt bulk update by user 1 including list2 (owned by user 2)
	// The transaction must fail and roll back list1's title!
	_, err = todolistService.UpdateBulk(ctx, 1, []db.Todolist{
		{ID: list1.ID, Title: "List 1 Mutated"},
		{ID: list2.ID, Title: "List 2 Hijacked"},
	})
	if !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound when attempting to update other user's list, got: %v", err)
	}

	// Verify rollback: list1 must remain "List 1 Original"
	persisted, err := todolistService.GetByID(ctx, list1.ID, 1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if persisted.Title != "List 1 Original" {
		t.Fatalf("Expected title to be rolled back to 'List 1 Original', got: %q", persisted.Title)
	}
}

func TestTodoService_TransactionRollback(t *testing.T) {
	ctx := context.Background()
	_, _, todoService := setupIsolatedServiceDB(t)

	// Create initial todos for user 1
	created, err := todoService.CreateBulk(ctx, 1, []db.Todo{
		{Message: "Todo 1", Check: false},
		{Message: "Todo 2", Check: false},
	})
	if err != nil || len(created) != 2 {
		t.Fatalf("CreateBulk failed: err=%v, len=%d", err, len(created))
	}

	newMsg := "Todo 1 Mutated"
	// Attempt bulk update with valid item and non-existent item
	// Transaction must roll back!
	_, err = todoService.UpdateBulk(ctx, 1, []db.TodoUpdateItem{
		{ID: created[0].ID, Message: &newMsg},
		{ID: 99999, Message: &newMsg},
	})
	if !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound on non-existent item in bulk update, got: %v", err)
	}

	// Verify rollback
	persisted, err := todoService.GetByID(ctx, created[0].ID, 1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if persisted.Message != "Todo 1" {
		t.Fatalf("Expected message to be rolled back to 'Todo 1', got: %q", persisted.Message)
	}
}

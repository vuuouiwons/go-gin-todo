package tests

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"go-gin-todo/internal/app"
	"go-gin-todo/internal/infra"
)

func setupIsolatedTestDB(t *testing.T) (*infra.Database, app.UserRepository, app.TodolistRepository, app.TodoRepository) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "integration.db")
	gormDB, err := infra.NewSQLiteDatabase(dbPath)
	if err != nil {
		t.Fatalf("failed to create isolated test DB: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := gormDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return infra.NewDatabase(gormDB),
		infra.NewUserRepository(gormDB),
		infra.NewTodolistRepository(gormDB),
		infra.NewTodoRepository(gormDB)
}

func TestIntegration_UserRepositoryConstraints(t *testing.T) {
	ctx := context.Background()
	_, userRepo, _, _ := setupIsolatedTestDB(t)

	u := &app.User{Email: "alice@example.com", Password: "hashedpassword"}
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if u.ID == 0 {
		t.Fatalf("expected auto-incremented ID")
	}

	// Duplicate email violates unique index
	dup := &app.User{Email: "alice@example.com", Password: "otherpassword"}
	if err := userRepo.Create(ctx, dup); err == nil {
		t.Fatalf("expected unique constraint error on duplicate email")
	}

	got, err := userRepo.GetByEmail(ctx, "alice@example.com")
	if err != nil || got.ID != u.ID {
		t.Fatalf("GetByEmail failed: got=%+v err=%v", got, err)
	}

	if _, err := userRepo.GetByEmail(ctx, "missing@example.com"); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("expected app.ErrNotFound for missing email, got %v", err)
	}
}

func TestIntegration_TodolistQueriesAndTransactionRollback(t *testing.T) {
	ctx := context.Background()
	dbAdapter, _, listRepo, _ := setupIsolatedTestDB(t)

	// Seed data for user 1 and user 2
	_, err := listRepo.CreateBulk(ctx, []app.Todolist{
		{UserID: 1, Title: "Alpha Project"},
		{UserID: 1, Title: "Beta Groceries"},
		{UserID: 1, Title: "Gamma Project"},
		{UserID: 2, Title: "User2 Project"},
	})
	if err != nil {
		t.Fatalf("seed CreateBulk failed: %v", err)
	}

	queries := app.NewTodolistQueries(dbAdapter, listRepo)

	// Query pagination & ordering (id DESC)
	page1, meta1, err := queries.GetBulk(ctx, 1, "", 1, 2)
	if err != nil {
		t.Fatalf("GetBulk page 1 failed: %v", err)
	}
	if len(page1) != 2 || meta1.TotalItems != 3 || meta1.TotalPages != 2 || !meta1.HasNext || meta1.HasPrev {
		t.Fatalf("unexpected page1 result/meta: len=%d meta=%+v", len(page1), meta1)
	}
	if page1[0].Title != "Gamma Project" {
		t.Fatalf("expected newest item first (id DESC), got %q", page1[0].Title)
	}

	// Query search filter via clause.Like (no raw SQL)
	filtered, metaFiltered, err := queries.GetBulk(ctx, 1, "Project", 1, 10)
	if err != nil {
		t.Fatalf("GetBulk search failed: %v", err)
	}
	if len(filtered) != 2 || metaFiltered.TotalItems != 2 {
		t.Fatalf("expected 2 matching Project items for user 1, got %d", len(filtered))
	}

	// Real DB transaction rollback test (and Propagation.Required) via Transactor + TodolistCommands
	commands := app.NewTodolistCommands(listRepo)
	err = dbAdapter.ExecInTransaction(ctx, func(txCtx context.Context) error {
		return dbAdapter.ExecInTransaction(txCtx, func(nestedCtx context.Context) error {
			_, err := commands.UpdateBulk(nestedCtx, 1, []app.Todolist{
				{ID: filtered[0].ID, Title: "Mutated Then Rolled Back"},
				{ID: 99999, Title: "Nonexistent"},
			})
			return err
		})
	})
	if !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	persisted, err := queries.GetByID(ctx, filtered[0].ID, 1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if persisted.Title == "Mutated Then Rolled Back" {
		t.Fatalf("expected DB transaction rollback to revert title change, but change persisted")
	}
}

func TestIntegration_TodoQueriesAndFilters(t *testing.T) {
	ctx := context.Background()
	dbAdapter, _, _, todoRepo := setupIsolatedTestDB(t)

	_, err := todoRepo.CreateBulk(ctx, []app.Todo{
		{UserID: 1, Message: "Buy organic milk", Check: false},
		{UserID: 1, Message: "Write integration tests", Check: true},
		{UserID: 1, Message: "Review PRs", Check: true},
		{UserID: 2, Message: "Buy milk for user 2", Check: false},
	})
	if err != nil {
		t.Fatalf("seed todos failed: %v", err)
	}

	queries := app.NewTodoQueries(dbAdapter, todoRepo)

	// Filter check=true
	onlyDone := true
	doneItems, doneMeta, err := queries.GetBulk(ctx, 1, "", &onlyDone, 1, 10)
	if err != nil {
		t.Fatalf("GetBulk check=true failed: %v", err)
	}
	if len(doneItems) != 2 || doneMeta.TotalItems != 2 {
		t.Fatalf("expected 2 completed todos for user 1, got %d", len(doneItems))
	}

	// Filter check=false + search="milk"
	onlyOpen := false
	openMilk, milkMeta, err := queries.GetBulk(ctx, 1, "milk", &onlyOpen, 1, 10)
	if err != nil {
		t.Fatalf("GetBulk check=false search=milk failed: %v", err)
	}
	if len(openMilk) != 1 || milkMeta.TotalItems != 1 || openMilk[0].Message != "Buy organic milk" {
		t.Fatalf("unexpected filtered todos: %+v", openMilk)
	}

	// Verify DB column name user_id works in delete bulk
	deleted, err := todoRepo.DeleteBulk(ctx, []uint{doneItems[0].ID, doneItems[1].ID}, 1)
	if err != nil || deleted != 2 {
		t.Fatalf("expected 2 deleted, got %d err=%v", deleted, err)
	}
}

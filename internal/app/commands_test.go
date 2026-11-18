package app

import (
	"context"
	"errors"
	"testing"
)

// --- In-memory Mocks (no database spun up) ---

type mockTransactor struct {
	calls    int
	rollback func()
	snapshot func()
}

func (m *mockTransactor) ExecInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.calls++
	if m.snapshot != nil {
		m.snapshot()
	}
	if err := fn(ctx); err != nil {
		if m.rollback != nil {
			m.rollback()
		}
		return err
	}
	return nil
}

type mockHasher struct {
	failHash bool
}

func (m *mockHasher) Hash(password string) (string, error) {
	if m.failHash {
		return "", errors.New("hash error")
	}
	return "hashed:" + password, nil
}

func (m *mockHasher) Compare(hashedPassword, password string) bool {
	return hashedPassword == "hashed:"+password
}

type mockTokenManager struct {
	failGen bool
}

func (m *mockTokenManager) Generate(userID uint, email string) (string, error) {
	if m.failGen {
		return "", errors.New("token error")
	}
	return "token:" + email, nil
}

func (m *mockTokenManager) Validate(tokenStr string) (*TokenClaims, error) {
	return nil, ErrUnauthorized
}

type mockUserRepo struct {
	users      map[uint]User
	nextID     uint
	failCreate bool
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[uint]User), nextID: 1}
}

func (m *mockUserRepo) Create(_ context.Context, user *User) error {
	if m.failCreate {
		return errors.New("db failure")
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = *user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id uint) (*User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &u, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*User, error) {
	for _, u := range m.users {
		if u.Email == email {
			copyU := u
			return &copyU, nil
		}
	}
	return nil, ErrNotFound
}

type mockTodolistRepo struct {
	items    map[uint]Todolist
	backup   map[uint]Todolist
	nextID   uint
	failSave bool
}

func newMockTodolistRepo() *mockTodolistRepo {
	return &mockTodolistRepo{items: make(map[uint]Todolist), nextID: 1}
}

func (m *mockTodolistRepo) snapshot() {
	m.backup = make(map[uint]Todolist, len(m.items))
	for k, v := range m.items {
		m.backup[k] = v
	}
}

func (m *mockTodolistRepo) rollback() {
	if m.backup != nil {
		m.items = m.backup
	}
}

func (m *mockTodolistRepo) Create(_ context.Context, todolist *Todolist) error {
	if m.failSave {
		return errors.New("db error")
	}
	todolist.ID = m.nextID
	m.nextID++
	m.items[todolist.ID] = *todolist
	return nil
}

func (m *mockTodolistRepo) CreateBulk(ctx context.Context, todolists []Todolist) ([]Todolist, error) {
	out := make([]Todolist, len(todolists))
	for i := range todolists {
		item := todolists[i]
		if err := m.Create(ctx, &item); err != nil {
			return nil, err
		}
		out[i] = item
	}
	return out, nil
}

func (m *mockTodolistRepo) GetByID(_ context.Context, id, userID uint) (*Todolist, error) {
	item, ok := m.items[id]
	if !ok || item.UserID != userID {
		return nil, ErrNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (m *mockTodolistRepo) Update(_ context.Context, todolist *Todolist) error {
	if m.failSave {
		return errors.New("db error")
	}
	m.items[todolist.ID] = *todolist
	return nil
}

func (m *mockTodolistRepo) DeleteByID(_ context.Context, id, userID uint) (int64, error) {
	item, ok := m.items[id]
	if !ok || item.UserID != userID {
		return 0, nil
	}
	delete(m.items, id)
	return 1, nil
}

func (m *mockTodolistRepo) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	var count int64
	for _, id := range ids {
		n, err := m.DeleteByID(ctx, id, userID)
		if err != nil {
			return 0, err
		}
		count += n
	}
	return count, nil
}

type mockTodoRepo struct {
	items    map[uint]Todo
	backup   map[uint]Todo
	nextID   uint
	failSave bool
}

func newMockTodoRepo() *mockTodoRepo {
	return &mockTodoRepo{items: make(map[uint]Todo), nextID: 1}
}

func (m *mockTodoRepo) snapshot() {
	m.backup = make(map[uint]Todo, len(m.items))
	for k, v := range m.items {
		m.backup[k] = v
	}
}

func (m *mockTodoRepo) rollback() {
	if m.backup != nil {
		m.items = m.backup
	}
}

func (m *mockTodoRepo) Create(_ context.Context, todo *Todo) error {
	if m.failSave {
		return errors.New("db error")
	}
	todo.ID = m.nextID
	m.nextID++
	m.items[todo.ID] = *todo
	return nil
}

func (m *mockTodoRepo) CreateBulk(ctx context.Context, todos []Todo) ([]Todo, error) {
	out := make([]Todo, len(todos))
	for i := range todos {
		item := todos[i]
		if err := m.Create(ctx, &item); err != nil {
			return nil, err
		}
		out[i] = item
	}
	return out, nil
}

func (m *mockTodoRepo) GetByID(_ context.Context, id, userID uint) (*Todo, error) {
	item, ok := m.items[id]
	if !ok || item.UserID != userID {
		return nil, ErrNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (m *mockTodoRepo) Update(_ context.Context, todo *Todo) error {
	if m.failSave {
		return errors.New("db error")
	}
	m.items[todo.ID] = *todo
	return nil
}

func (m *mockTodoRepo) DeleteByID(_ context.Context, id, userID uint) (int64, error) {
	item, ok := m.items[id]
	if !ok || item.UserID != userID {
		return 0, nil
	}
	delete(m.items, id)
	return 1, nil
}

func (m *mockTodoRepo) DeleteBulk(ctx context.Context, ids []uint, userID uint) (int64, error) {
	var count int64
	for _, id := range ids {
		n, err := m.DeleteByID(ctx, id, userID)
		if err != nil {
			return 0, err
		}
		count += n
	}
	return count, nil
}

// --- Unit Tests for Application Commands ---

func TestAuthCommands(t *testing.T) {
	ctx := context.Background()
	userRepo := newMockUserRepo()
	hasher := &mockHasher{}
	tm := &mockTokenManager{}
	cmd := NewAuthCommands(userRepo, hasher, tm)

	t.Run("Register invalid input", func(t *testing.T) {
		if _, err := cmd.Register(ctx, "  ", "secret123"); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for empty email, got %v", err)
		}
		if _, err := cmd.Register(ctx, "user@test.com", "123"); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for short password, got %v", err)
		}
	})

	t.Run("Register normalizes email and hashes password", func(t *testing.T) {
		u, err := cmd.Register(ctx, "  Alice@Example.COM ", "secret123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Email != "alice@example.com" || u.Password != "hashed:secret123" {
			t.Fatalf("unexpected user stored: %+v", u)
		}
	})

	t.Run("Register duplicate returns ErrConflict", func(t *testing.T) {
		if _, err := cmd.Register(ctx, "alice@example.com", "secret123"); !errors.Is(err, ErrConflict) {
			t.Fatalf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("Login success and failure paths", func(t *testing.T) {
		user, token, err := cmd.Login(ctx, " ALICE@example.com ", "secret123")
		if err != nil || user == nil || token != "token:alice@example.com" {
			t.Fatalf("expected valid login, got user=%v token=%s err=%v", user, token, err)
		}

		if _, _, err := cmd.Login(ctx, "alice@example.com", "wrong"); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized on wrong password, got %v", err)
		}
		if _, _, err := cmd.Login(ctx, "nobody@example.com", "secret123"); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized on missing user, got %v", err)
		}
		if _, _, err := cmd.Login(ctx, "", "secret123"); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized on empty email, got %v", err)
		}
	})
}

func TestTodolistCommands(t *testing.T) {
	ctx := context.Background()
	repo := newMockTodolistRepo()
	tx := &mockTransactor{snapshot: repo.snapshot, rollback: repo.rollback}
	cmd := NewTodolistCommands(repo)

	t.Run("CreateOne and validation", func(t *testing.T) {
		if _, err := cmd.CreateOne(ctx, 1, "   "); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest, got %v", err)
		}
		created, err := cmd.CreateOne(ctx, 1, "  Work  ")
		if err != nil || created.Title != "Work" || created.UserID != 1 {
			t.Fatalf("unexpected CreateOne result: %+v, err=%v", created, err)
		}
	})

	t.Run("CreateBulk and validation", func(t *testing.T) {
		if _, err := cmd.CreateBulk(ctx, 1, nil); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for empty slice, got %v", err)
		}
		if _, err := cmd.CreateBulk(ctx, 1, []string{"Valid", "  "}); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for blank title, got %v", err)
		}
		created, err := cmd.CreateBulk(ctx, 1, []string{" Home ", " Gym "})
		if err != nil || len(created) != 2 || created[0].Title != "Home" || created[1].Title != "Gym" {
			t.Fatalf("unexpected CreateBulk result: %+v, err=%v", created, err)
		}
	})

	t.Run("UpdateByID respects ownership", func(t *testing.T) {
		if _, err := cmd.UpdateByID(ctx, 1, 2, "Hacked"); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound for non-owner, got %v", err)
		}
		updated, err := cmd.UpdateByID(ctx, 1, 1, "  Work Updated ")
		if err != nil || updated.Title != "Work Updated" {
			t.Fatalf("unexpected UpdateByID result: %+v, err=%v", updated, err)
		}
	})

	t.Run("UpdateBulk propagates error and rolls back inside Transactor", func(t *testing.T) {
		// Create item owned by user 2
		other, _ := cmd.CreateOne(ctx, 2, "Other User List")

		// Attempt bulk update for user 1 inside transaction: ID 1 succeeds, then other.ID fails -> rollback keeps ID 1 unchanged
		err := tx.ExecInTransaction(ctx, func(txCtx context.Context) error {
			_, err := cmd.UpdateBulk(txCtx, 1, []Todolist{
				{ID: 1, Title: "Should Roll Back"},
				{ID: other.ID, Title: "Cannot Touch"},
			})
			return err
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound during bulk update, got %v", err)
		}
		if tx.calls != 1 {
			t.Fatalf("expected 1 transaction execution, got %d", tx.calls)
		}
		stored, _ := repo.GetByID(ctx, 1, 1)
		if stored.Title != "Work Updated" {
			t.Fatalf("expected rollback to preserve 'Work Updated', got %q", stored.Title)
		}

		// Valid bulk update commits
		updated, err := cmd.UpdateBulk(ctx, 1, []Todolist{
			{ID: 1, Title: "Work Final"},
			{ID: 2, Title: "Home Final"},
		})
		if err != nil || len(updated) != 2 || updated[0].Title != "Work Final" {
			t.Fatalf("unexpected bulk update commit result: %+v, err=%v", updated, err)
		}
	})

	t.Run("DeleteByID and DeleteBulk", func(t *testing.T) {
		if _, err := cmd.DeleteByID(ctx, 0, 1); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest, got %v", err)
		}
		if _, err := cmd.DeleteBulk(ctx, nil, 1); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest, got %v", err)
		}
		n, err := cmd.DeleteByID(ctx, 1, 1)
		if err != nil || n != 1 {
			t.Fatalf("expected 1 deleted, got %d, err=%v", n, err)
		}
		nBulk, err := cmd.DeleteBulk(ctx, []uint{2, 3, 999}, 1)
		if err != nil || nBulk != 2 {
			t.Fatalf("expected 2 bulk deleted, got %d, err=%v", nBulk, err)
		}
	})
}

func TestTodoCommands(t *testing.T) {
	ctx := context.Background()
	repo := newMockTodoRepo()
	tx := &mockTransactor{snapshot: repo.snapshot, rollback: repo.rollback}
	cmd := NewTodoCommands(repo)

	t.Run("CreateOne and CreateBulk", func(t *testing.T) {
		if _, err := cmd.CreateOne(ctx, 1, "  ", false); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for empty message, got %v", err)
		}
		one, err := cmd.CreateOne(ctx, 1, " Buy milk ", false)
		if err != nil || one.Message != "Buy milk" || one.Check {
			t.Fatalf("unexpected CreateOne result: %+v, err=%v", one, err)
		}

		bulk, err := cmd.CreateBulk(ctx, 1, []Todo{
			{Message: " Task 2 ", Check: true},
			{Message: " Task 3 ", Check: false},
		})
		if err != nil || len(bulk) != 2 || bulk[0].Message != "Task 2" || !bulk[0].Check {
			t.Fatalf("unexpected CreateBulk result: %+v, err=%v", bulk, err)
		}
	})

	t.Run("UpdateByID partial fields and validation", func(t *testing.T) {
		if _, err := cmd.UpdateByID(ctx, 1, 1, nil, nil); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest when no update fields provided, got %v", err)
		}
		blank := "   "
		if _, err := cmd.UpdateByID(ctx, 1, 1, &blank, nil); !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for blank message, got %v", err)
		}

		done := true
		updated, err := cmd.UpdateByID(ctx, 1, 1, nil, &done)
		if err != nil || !updated.Check || updated.Message != "Buy milk" {
			t.Fatalf("unexpected UpdateByID result: %+v, err=%v", updated, err)
		}
	})

	t.Run("UpdateBulk propagates error and rolls back inside Transactor", func(t *testing.T) {
		msg := "Updated Task 1"
		falseVal := false
		err := tx.ExecInTransaction(ctx, func(txCtx context.Context) error {
			_, err := cmd.UpdateBulk(txCtx, 1, []TodoUpdateItem{
				{ID: 1, Message: &msg, Check: &falseVal},
				{ID: 999, Check: &falseVal},
			})
			return err
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
		if tx.calls != 1 {
			t.Fatalf("expected transaction call, got %d", tx.calls)
		}
		stored, _ := repo.GetByID(ctx, 1, 1)
		if stored.Message != "Buy milk" || !stored.Check {
			t.Fatalf("expected rollback to preserve previous state, got %+v", stored)
		}

		// Commit succeeds
		res, err := cmd.UpdateBulk(ctx, 1, []TodoUpdateItem{
			{ID: 1, Message: &msg, Check: &falseVal},
		})
		if err != nil || len(res) != 1 || res[0].Message != "Updated Task 1" || res[0].Check {
			t.Fatalf("expected committed update, got %+v, err=%v", res, err)
		}
	})

	t.Run("DeleteByID and DeleteBulk", func(t *testing.T) {
		n, err := cmd.DeleteByID(ctx, 1, 1)
		if err != nil || n != 1 {
			t.Fatalf("expected 1 deleted, got %d, err=%v", n, err)
		}
		nBulk, err := cmd.DeleteBulk(ctx, []uint{2, 3}, 1)
		if err != nil || nBulk != 2 {
			t.Fatalf("expected 2 bulk deleted, got %d, err=%v", nBulk, err)
		}
	})
}

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"go-gin-todo/src/db"
	"go-gin-todo/src/db/repository"
	"go-gin-todo/src/routes"
	"go-gin-todo/src/services"
	"go-gin-todo/src/utils"
	"gorm.io/gorm"
)

var (
	testRouter *gin.Engine
	testDB     *gorm.DB
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	testDBPath := "test_e2e_todo.db"
	_ = os.Remove(testDBPath)

	cfg := &utils.Config{
		Port:              "8080",
		JWTSecret:         "test-secret-key-12345",
		JWTExpiresInHours: 2,
		DBPath:            testDBPath,
		GinMode:           "test",
	}

	database, err := db.Connect(cfg.DBPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}
	testDB = database

	userRepo := repository.NewUserRepository(database)
	todolistRepo := repository.NewTodolistRepository(database)
	todoRepo := repository.NewTodoRepository(database)

	authService := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiresInHours)
	todolistService := services.NewTodolistService(database, todolistRepo)
	todoService := services.NewTodoService(database, todoRepo)

	authHandler := routes.NewAuthHandler(authService)
	todolistHandler := routes.NewTodolistHandler(todolistService)
	todoHandler := routes.NewTodoHandler(todoService)

	testRouter = routes.SetupRouter(cfg.GinMode, cfg.JWTSecret, authHandler, todolistHandler, todoHandler)

	code := m.Run()

	sqlDB, err := testDB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
	_ = os.Remove(testDBPath)

	os.Exit(code)
}

func performRequest(r http.Handler, method, path string, body []byte, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func registerAndLogin(t *testing.T, email, password string) string {
	// Register
	regPayload := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	wReg := performRequest(testRouter, "POST", "/api/v1/register", []byte(regPayload), "")
	if wReg.Code != http.StatusCreated && wReg.Code != http.StatusConflict {
		t.Fatalf("Failed to register test user: %s", wReg.Body.String())
	}

	// Login
	loginPayload := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	wLogin := performRequest(testRouter, "POST", "/api/v1/login", []byte(loginPayload), "")
	if wLogin.Code != http.StatusOK {
		t.Fatalf("Failed to login test user: %s", wLogin.Body.String())
	}

	var res struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(wLogin.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	return res.Token
}

// 1. Test Auth: Register, Login, Duplicate email, Invalid password
func TestAuth(t *testing.T) {
	regBody := `{"email":"auth_clean@example.com","password":"password123"}`
	w := performRequest(testRouter, "POST", "/api/v1/register", []byte(regBody), "")
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Duplicate registration should fail with 409
	wDup := performRequest(testRouter, "POST", "/api/v1/register", []byte(regBody), "")
	if wDup.Code != http.StatusConflict {
		t.Errorf("Expected status 409 for duplicate email, got %d", wDup.Code)
	}

	// Login with correct credentials
	loginBody := `{"email":"auth_clean@example.com","password":"password123"}`
	wLog := performRequest(testRouter, "POST", "/api/v1/login", []byte(loginBody), "")
	if wLog.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", wLog.Code, wLog.Body.String())
	}

	// Login with wrong password
	wrongBody := `{"email":"auth_clean@example.com","password":"wrongpassword"}`
	wWrong := performRequest(testRouter, "POST", "/api/v1/login", []byte(wrongBody), "")
	if wWrong.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong password, got %d", wWrong.Code)
	}
}

// 2. Test URI Versioning & Default Versioning (default version is 1)
func TestURIVersioning(t *testing.T) {
	token := registerAndLogin(t, "version_clean@example.com", "pass1234")

	// Explicit /api/v1/todolist
	wV1 := performRequest(testRouter, "GET", "/api/v1/todolist", nil, token)
	if wV1.Code != http.StatusOK {
		t.Errorf("Expected status 200 on /api/v1/todolist, got %d", wV1.Code)
	}

	// Default version mapping: /api/todolist
	wDefault := performRequest(testRouter, "GET", "/api/todolist", nil, token)
	if wDefault.Code != http.StatusOK {
		t.Errorf("Expected status 200 on default /api/todolist, got %d", wDefault.Code)
	}
}

// Test Swagger UI and OpenAPI JSON serving
func TestSwaggerAndOpenAPI(t *testing.T) {
	// Test /openapi.json endpoint
	wOpenAPI := performRequest(testRouter, "GET", "/openapi.json", nil, "")
	if wOpenAPI.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on /openapi.json, got %d", wOpenAPI.Code)
	}
	var spec struct {
		OpenAPI string `json:"openapi"`
		Info    struct {
			Title string `json:"title"`
		} `json:"info"`
	}
	if err := json.Unmarshal(wOpenAPI.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse openapi.json response: %v", err)
	}
	if spec.OpenAPI == "" || spec.Info.Title == "" {
		t.Fatalf("openapi.json missing expected fields: %+v", spec)
	}

	// Test /docs endpoint (Swagger UI)
	wSwagger := performRequest(testRouter, "GET", "/docs", nil, "")
	if wSwagger.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on /docs, got %d", wSwagger.Code)
	}
}

// 3. Test Todolist CRUD: Single & Bulk operations
func TestTodolistCRUD(t *testing.T) {
	token := registerAndLogin(t, "todolist_clean@example.com", "pass1234")

	// POST /todolist (single)
	postOneBody := `{"title":"My First Clean List"}`
	wPostOne := performRequest(testRouter, "POST", "/api/v1/todolist", []byte(postOneBody), token)
	if wPostOne.Code != http.StatusCreated {
		t.Fatalf("POST one todolist failed: %s", wPostOne.Body.String())
	}
	var createdOne struct {
		Data struct {
			ID     uint   `json:"id"`
			UserID uint   `json:"userId"`
			Title  string `json:"title"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wPostOne.Body.Bytes(), &createdOne)
	singleID := createdOne.Data.ID
	if singleID == 0 || createdOne.Data.Title != "My First Clean List" {
		t.Fatalf("Unexpected single todolist created: %+v", createdOne)
	}

	// POST /todolist (bulk)
	postBulkBody := `[{"title":"Work Projects"},{"title":"Groceries"}]`
	wPostBulk := performRequest(testRouter, "POST", "/api/v1/todolist", []byte(postBulkBody), token)
	if wPostBulk.Code != http.StatusCreated {
		t.Fatalf("POST bulk todolist failed: %s", wPostBulk.Body.String())
	}
	var createdBulk struct {
		Count int `json:"count"`
		Data  []struct {
			ID    uint   `json:"id"`
			Title string `json:"title"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wPostBulk.Body.Bytes(), &createdBulk)
	if createdBulk.Count != 2 || len(createdBulk.Data) != 2 {
		t.Fatalf("Expected 2 bulk created items, got %d", createdBulk.Count)
	}
	bulkID1 := createdBulk.Data[0].ID
	bulkID2 := createdBulk.Data[1].ID

	// GET (bulk): /todolist with pagination
	wGetBulk := performRequest(testRouter, "GET", "/api/v1/todolist?page=1&limit=2", nil, token)
	if wGetBulk.Code != http.StatusOK {
		t.Fatalf("GET bulk todolist failed: %s", wGetBulk.Body.String())
	}
	var listPageResp struct {
		Data       []struct{ ID uint } `json:"data"`
		Pagination struct {
			Page       int   `json:"page"`
			Limit      int   `json:"limit"`
			TotalItems int64 `json:"totalItems"`
			TotalPages int   `json:"totalPages"`
			HasNext    bool  `json:"hasNext"`
			HasPrev    bool  `json:"hasPrev"`
		} `json:"pagination"`
	}
	_ = json.Unmarshal(wGetBulk.Body.Bytes(), &listPageResp)
	if listPageResp.Pagination.Page != 1 || listPageResp.Pagination.Limit != 2 || listPageResp.Pagination.TotalItems != 3 {
		t.Fatalf("Unexpected todolist pagination metadata: %+v", listPageResp.Pagination)
	}
	if len(listPageResp.Data) != 2 || !listPageResp.Pagination.HasNext {
		t.Fatalf("Expected 2 items on page 1 with HasNext=true, got %+v", listPageResp)
	}

	// GET (one): /todolist/:id
	wGetOne := performRequest(testRouter, "GET", fmt.Sprintf("/api/v1/todolist/%d", singleID), nil, token)
	if wGetOne.Code != http.StatusOK {
		t.Fatalf("GET one todolist failed: %s", wGetOne.Body.String())
	}

	// PATCH (update by id via body on /todolist):
	patchOneBody := fmt.Sprintf(`{"id":%d,"title":"My First Clean List Updated"}`, singleID)
	wPatchOne := performRequest(testRouter, "PATCH", "/api/v1/todolist", []byte(patchOneBody), token)
	if wPatchOne.Code != http.StatusOK {
		t.Fatalf("PATCH one todolist failed: %s", wPatchOne.Body.String())
	}

	// PATCH (update by id via URL on /todolist/:id):
	patchByIDBody := `{"title":"Updated via ID URL"}`
	wPatchByID := performRequest(testRouter, "PATCH", fmt.Sprintf("/api/v1/todolist/%d", singleID), []byte(patchByIDBody), token)
	if wPatchByID.Code != http.StatusOK {
		t.Fatalf("PATCH todolist by ID url failed: %s", wPatchByID.Body.String())
	}

	// PATCH (bulk update):
	patchBulkBody := fmt.Sprintf(`[{"id":%d,"title":"Work Projects v2"},{"id":%d,"title":"Groceries v2"}]`, bulkID1, bulkID2)
	wPatchBulk := performRequest(testRouter, "PATCH", "/api/v1/todolist", []byte(patchBulkBody), token)
	if wPatchBulk.Code != http.StatusOK {
		t.Fatalf("PATCH bulk todolist failed: %s", wPatchBulk.Body.String())
	}

	// DELETE (single via body on /todolist):
	deleteSingleBody := fmt.Sprintf(`{"id":%d}`, singleID)
	wDeleteOne := performRequest(testRouter, "DELETE", "/api/v1/todolist", []byte(deleteSingleBody), token)
	if wDeleteOne.Code != http.StatusOK {
		t.Fatalf("DELETE one todolist via body failed: %s", wDeleteOne.Body.String())
	}

	// DELETE without body (query params only) must fail with 400 Bad Request
	wDeleteNoBody := performRequest(testRouter, "DELETE", fmt.Sprintf("/api/v1/todolist?id=%d", bulkID1), nil, token)
	if wDeleteNoBody.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request when DELETE /todolist has no body, got %d", wDeleteNoBody.Code)
	}

	// DELETE (bulk via body on /todolist):
	deleteBulkBody := fmt.Sprintf(`{"ids":[%d,%d]}`, bulkID1, bulkID2)
	wDeleteBulk := performRequest(testRouter, "DELETE", "/api/v1/todolist", []byte(deleteBulkBody), token)
	if wDeleteBulk.Code != http.StatusOK {
		t.Fatalf("DELETE bulk todolist via body failed: %s", wDeleteBulk.Body.String())
	}
}

// 4. Test Todo CRUD: Single & Bulk operations
func TestTodoCRUD(t *testing.T) {
	token := registerAndLogin(t, "todo_clean@example.com", "pass1234")

	// POST /todo (single)
	postOneBody := `{"message":"Finish Clean Architecture","check":false}`
	wPostOne := performRequest(testRouter, "POST", "/api/v1/todo", []byte(postOneBody), token)
	if wPostOne.Code != http.StatusCreated {
		t.Fatalf("POST one todo failed: %s", wPostOne.Body.String())
	}
	var createdOne struct {
		Data struct {
			ID      uint   `json:"id"`
			UserID  uint   `json:"userId"`
			Message string `json:"message"`
			Check   bool   `json:"check"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wPostOne.Body.Bytes(), &createdOne)
	singleID := createdOne.Data.ID
	if singleID == 0 || createdOne.Data.Message != "Finish Clean Architecture" || createdOne.Data.Check != false {
		t.Fatalf("Unexpected single todo created: %+v", createdOne)
	}

	// POST /todo (bulk)
	postBulkBody := `[{"message":"Buy coffee","check":false},{"message":"Submit PR","check":true}]`
	wPostBulk := performRequest(testRouter, "POST", "/api/v1/todo", []byte(postBulkBody), token)
	if wPostBulk.Code != http.StatusCreated {
		t.Fatalf("POST bulk todo failed: %s", wPostBulk.Body.String())
	}
	var createdBulk struct {
		Count int `json:"count"`
		Data  []struct {
			ID      uint   `json:"id"`
			Message string `json:"message"`
			Check   bool   `json:"check"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wPostBulk.Body.Bytes(), &createdBulk)
	if createdBulk.Count != 2 || len(createdBulk.Data) != 2 {
		t.Fatalf("Expected 2 bulk created todos, got %d", createdBulk.Count)
	}
	bulkID1 := createdBulk.Data[0].ID
	bulkID2 := createdBulk.Data[1].ID

	// GET (bulk): /todo with pagination
	wGetBulk := performRequest(testRouter, "GET", "/api/v1/todo?page=1&limit=2", nil, token)
	if wGetBulk.Code != http.StatusOK {
		t.Fatalf("GET bulk todo failed: %s", wGetBulk.Body.String())
	}
	var todoPageResp struct {
		Data       []struct{ ID uint } `json:"data"`
		Pagination struct {
			Page       int   `json:"page"`
			Limit      int   `json:"limit"`
			TotalItems int64 `json:"totalItems"`
			TotalPages int   `json:"totalPages"`
			HasNext    bool  `json:"hasNext"`
			HasPrev    bool  `json:"hasPrev"`
		} `json:"pagination"`
	}
	_ = json.Unmarshal(wGetBulk.Body.Bytes(), &todoPageResp)
	if todoPageResp.Pagination.Page != 1 || todoPageResp.Pagination.Limit != 2 || todoPageResp.Pagination.TotalItems != 3 {
		t.Fatalf("Unexpected todo pagination metadata: %+v", todoPageResp.Pagination)
	}
	if len(todoPageResp.Data) != 2 || !todoPageResp.Pagination.HasNext {
		t.Fatalf("Expected 2 items on page 1 with HasNext=true, got %+v", todoPageResp)
	}

	// GET (one): /todo/:id
	wGetOne := performRequest(testRouter, "GET", fmt.Sprintf("/api/v1/todo/%d", singleID), nil, token)
	if wGetOne.Code != http.StatusOK {
		t.Fatalf("GET one todo failed: %s", wGetOne.Body.String())
	}

	// PATCH (update by id via body on /todo):
	patchOneBody := fmt.Sprintf(`{"id":%d,"message":"Finish Clean Architecture v2","check":true}`, singleID)
	wPatchOne := performRequest(testRouter, "PATCH", "/api/v1/todo", []byte(patchOneBody), token)
	if wPatchOne.Code != http.StatusOK {
		t.Fatalf("PATCH one todo failed: %s", wPatchOne.Body.String())
	}

	// PATCH (update by id via URL on /todo/:id):
	patchByIDBody := `{"check":false}`
	wPatchByID := performRequest(testRouter, "PATCH", fmt.Sprintf("/api/v1/todo/%d", singleID), []byte(patchByIDBody), token)
	if wPatchByID.Code != http.StatusOK {
		t.Fatalf("PATCH todo by ID url failed: %s", wPatchByID.Body.String())
	}

	// PATCH (bulk update):
	patchBulkBody := fmt.Sprintf(`[{"id":%d,"check":true},{"id":%d,"message":"Submit PR ASAP","check":false}]`, bulkID1, bulkID2)
	wPatchBulk := performRequest(testRouter, "PATCH", "/api/v1/todo", []byte(patchBulkBody), token)
	if wPatchBulk.Code != http.StatusOK {
		t.Fatalf("PATCH bulk todo failed: %s", wPatchBulk.Body.String())
	}

	// DELETE (single via body on /todo):
	deleteSingleBody := fmt.Sprintf(`{"id":%d}`, singleID)
	wDeleteOne := performRequest(testRouter, "DELETE", "/api/v1/todo", []byte(deleteSingleBody), token)
	if wDeleteOne.Code != http.StatusOK {
		t.Fatalf("DELETE one todo via body failed: %s", wDeleteOne.Body.String())
	}

	// DELETE without body (query params only) must fail with 400 Bad Request
	wDeleteNoBody := performRequest(testRouter, "DELETE", fmt.Sprintf("/api/v1/todo?id=%d", bulkID1), nil, token)
	if wDeleteNoBody.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request when DELETE /todo has no body, got %d", wDeleteNoBody.Code)
	}

	// DELETE (bulk via body on /todo):
	deleteBulkBody := fmt.Sprintf(`[%d,%d]`, bulkID1, bulkID2)
	wDeleteBulk := performRequest(testRouter, "DELETE", "/api/v1/todo", []byte(deleteBulkBody), token)
	if wDeleteBulk.Code != http.StatusOK {
		t.Fatalf("DELETE bulk todo via body failed: %s", wDeleteBulk.Body.String())
	}
}

// 5. Test User Data Isolation (User A cannot access User B's todos or todolists)
func TestUserDataIsolation(t *testing.T) {
	tokenA := registerAndLogin(t, "user_a_clean@example.com", "pass1234")
	tokenB := registerAndLogin(t, "user_b_clean@example.com", "pass1234")

	// User A creates a todolist and a todo
	wListA := performRequest(testRouter, "POST", "/api/v1/todolist", []byte(`{"title":"Secret List A"}`), tokenA)
	if wListA.Code != http.StatusCreated {
		t.Fatalf("Failed to create list for user A")
	}
	var resListA struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wListA.Body.Bytes(), &resListA)

	wTodoA := performRequest(testRouter, "POST", "/api/v1/todo", []byte(`{"message":"Secret Todo A"}`), tokenA)
	if wTodoA.Code != http.StatusCreated {
		t.Fatalf("Failed to create todo for user A")
	}
	var resTodoA struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wTodoA.Body.Bytes(), &resTodoA)

	// User B attempts to access User A's todolist -> should get 404
	wGetB := performRequest(testRouter, "GET", fmt.Sprintf("/api/v1/todolist/%d", resListA.Data.ID), nil, tokenB)
	if wGetB.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for User B accessing User A's todolist, got %d", wGetB.Code)
	}

	// User B attempts to update User A's todolist -> should get 404
	wPatchB := performRequest(testRouter, "PATCH", fmt.Sprintf("/api/v1/todolist/%d", resListA.Data.ID), []byte(`{"title":"Hacked"}`), tokenB)
	if wPatchB.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for User B updating User A's todolist, got %d", wPatchB.Code)
	}

	// User B attempts to delete User A's todo via body -> should delete 0 rows
	wDelBodyB := performRequest(testRouter, "DELETE", "/api/v1/todo", []byte(fmt.Sprintf(`{"id":%d}`, resTodoA.Data.ID)), tokenB)
	if wDelBodyB.Code != http.StatusOK {
		t.Errorf("Expected 200 for User B attempting delete on User A's todo, got %d", wDelBodyB.Code)
	}
	var delResult struct {
		DeletedCount int64 `json:"deletedCount"`
	}
	_ = json.Unmarshal(wDelBodyB.Body.Bytes(), &delResult)
	if delResult.DeletedCount != 0 {
		t.Errorf("User B should have deleted 0 rows, but deleted %d", delResult.DeletedCount)
	}

	// User B attempts to delete User A's todo via URL ID -> should get 404
	wDelB := performRequest(testRouter, "DELETE", fmt.Sprintf("/api/v1/todo/%d", resTodoA.Data.ID), nil, tokenB)
	if wDelB.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for User B deleting User A's todo, got %d", wDelB.Code)
	}

	// Verify transaction rollback over HTTP:
	// User B creates a list, then User A sends a bulk PATCH modifying both User A's list and User B's list.
	// Expect 404 Not Found AND User A's list must remain "Secret List A".
	wListB := performRequest(testRouter, "POST", "/api/v1/todolist", []byte(`{"title":"Secret List B"}`), tokenB)
	if wListB.Code != http.StatusCreated {
		t.Fatalf("Failed to create list for user B")
	}
	var resListB struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wListB.Body.Bytes(), &resListB)

	bulkPatchRollback := fmt.Sprintf(`[{"id":%d,"title":"Mutated A"},{"id":%d,"title":"Hacked B"}]`, resListA.Data.ID, resListB.Data.ID)
	wPatchRollback := performRequest(testRouter, "PATCH", "/api/v1/todolist", []byte(bulkPatchRollback), tokenA)
	if wPatchRollback.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for mixed-owner bulk PATCH, got %d", wPatchRollback.Code)
	}

	wVerifyA := performRequest(testRouter, "GET", fmt.Sprintf("/api/v1/todolist/%d", resListA.Data.ID), nil, tokenA)
	if wVerifyA.Code != http.StatusOK {
		t.Fatalf("Failed to fetch User A list after rollback: %d", wVerifyA.Code)
	}
	var verifyResp struct {
		Data struct {
			Title string `json:"title"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wVerifyA.Body.Bytes(), &verifyResp)
	if verifyResp.Data.Title != "Secret List A" {
		t.Errorf("Expected transaction rollback to preserve User A's title 'Secret List A', got %q", verifyResp.Data.Title)
	}
}

func TestTransactionLogsDemo(t *testing.T) {
	log.Println("--- 1. POST /api/v1/register (201 -> COMMIT) ---")
	wReg := performRequest(testRouter, "POST", "/api/v1/register", []byte(`{"email":"tx@test.com","password":"password123"}`), "")
	if wReg.Code != http.StatusCreated {
		t.Fatalf("expected 201 on register, got %d: %s", wReg.Code, wReg.Body.String())
	}

	wLogin := performRequest(testRouter, "POST", "/api/v1/login", []byte(`{"email":"tx@test.com","password":"password123"}`), "")
	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d: %s", wLogin.Code, wLogin.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)

	log.Println("--- 2. POST /api/v1/todo bulk create (201 -> COMMIT) ---")
	wBulkCreate := performRequest(testRouter, "POST", "/api/v1/todo", []byte(`[{"message":"Todo 1"},{"message":"Todo 2"}]`), loginResp.Token)
	if wBulkCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 on bulk create, got %d: %s", wBulkCreate.Code, wBulkCreate.Body.String())
	}
	var createResp struct {
		Data []struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wBulkCreate.Body.Bytes(), &createResp)
	firstID := createResp.Data[0].ID

	log.Println("--- 3. PATCH /api/v1/todo where 2nd item is missing (404 -> ROLLBACK) ---")
	patchPayload := fmt.Sprintf(`[{"id":%d,"message":"Todo 1 Mutated"},{"id":999,"message":"Missing"}]`, firstID)
	wPatch := performRequest(testRouter, "PATCH", "/api/v1/todo", []byte(patchPayload), loginResp.Token)
	if wPatch.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on missing 2nd item, got %d: %s", wPatch.Code, wPatch.Body.String())
	}
}

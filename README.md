# Go Gin Todo Backend Server (Clean Architecture)

A robust RESTful backend API built with Go, the [Gin Web Framework](https://github.com/gin-gonic/gin), and structured according to **Uncle Bob's Clean Architecture**.

It features JWT authentication, bcrypt password hashing, pure-Go SQLite persistence with GORM (no CGO required), URI versioning (`v1` as default), and full single/bulk CRUD capabilities for Todolists and Todos.

---

## Clean Architecture Layers

The codebase is organized following the Dependency Rule (dependencies point inward toward the Domain):

```text
go-gin-todo/
├── cmd/
│   └── server/
│       └── main.go                               # Production server entry point
├── internal/
│   ├── domain/                                   # 1. Enterprise Business Rules (Entities & Contracts)
│   │   ├── user.go                               # User entity & UserRepository interface
│   │   ├── todolist.go                           # Todolist entity & TodolistRepository interface
│   │   ├── todo.go                               # Todo entity & TodoRepository interface
│   │   └── errors.go                             # Standard domain error definitions
│   │
│   ├── usecase/                                  # 2. Application Business Rules (Use Cases)
│   │   ├── auth_usecase.go                       # Register & Login orchestration
│   │   ├── todolist_usecase.go                   # Todolist CRUD business logic
│   │   └── todo_usecase.go                       # Todo CRUD business logic
│   │
│   ├── repository/                               # 3. Interface Adapters - Persistence (GORM / SQLite)
│   │   ├── user_repository.go                    # GORM implementation of UserRepository
│   │   ├── todolist_repository.go                # GORM implementation of TodolistRepository
│   │   └── todo_repository.go                    # GORM implementation of TodoRepository
│   │
│   ├── security/                                 # 3. Infrastructure - Security Adapters
│   │   ├── password.go                           # PasswordHasher (bcrypt)
│   │   └── token.go                              # TokenManager (JWT)
│   │
│   ├── delivery/
│   │   └── http/                                 # 3. Interface Adapters - Delivery (Gin HTTP)
│   │       ├── dto/                              # Data Transfer Objects (Request/Response)
│   │       │   ├── auth_dto.go
│   │       │   ├── todolist_dto.go
│   │       │   └── todo_dto.go
│   │       ├── handler/                          # HTTP Handlers (Controllers)
│   │       │   ├── auth_handler.go
│   │       │   ├── todolist_handler.go
│   │       │   └── todo_handler.go
│   │       ├── middleware/                       # HTTP Middlewares (JWT Auth, CORS)
│   │       │   ├── auth.go
│   │       │   └── cors.go
│   │       └── router.go                         # Route definitions & URI versioning
│   │
│   ├── config/                                   # Application configuration & .env loader
│   │   └── config.go
│   ├── database/                                 # SQLite pure-Go connection & migrations
│   │   └── sqlite.go
│   └── app/                                      # 4. Dependency Injection / Composition Root
│       └── app.go                                # Wires repos, usecases, security, and handlers
├── tests/
│   └── api_test.go                               # Comprehensive integration test suite
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── main.go                                       # Root application entry point
└── README.md
```

---

## Database Tables

### 1. `Users`
| Column | Type | Attributes | Description |
|---|---|---|---|
| `id` | integer | Primary Key, Auto Increment | Unique User ID |
| `email` | varchar(255) | Unique, Not Null | User email address |
| `password` | varchar(255) | Not Null | Bcrypt hashed password |
| `created_at` | datetime | Auto | Timestamp |
| `updated_at` | datetime | Auto | Timestamp |

### 2. `Todolist`
| Column | Type | Attributes | Description |
|---|---|---|---|
| `id` | integer | Primary Key, Auto Increment | Unique List ID |
| `user_id` | integer | Index, Not Null | Owner's User ID |
| `title` | varchar(255) | Not Null | List title |
| `created_at` | datetime | Auto | Timestamp |
| `updated_at` | datetime | Auto | Timestamp |

### 3. `Todo`
| Column | Type | Attributes | Description |
|---|---|---|---|
| `id` | integer | Primary Key, Auto Increment | Unique Todo ID |
| `user_id` | integer | Index, Not Null | Owner's User ID |
| `message` | text | Not Null | Task description |
| `check` | boolean | Default: `false` | Completion status |
| `created_at` | datetime | Auto | Timestamp |
| `updated_at` | datetime | Auto | Timestamp |

---

## Getting Started

### Prerequisites
- Go 1.20+ (compiled and tested with Go 1.27 on Windows)
- Zero CGO required (uses pure Go SQLite driver `github.com/glebarez/sqlite`)

### Run Server
```bash
go run .
```
or
```bash
go run cmd/server/main.go
```

The server listens on `http://localhost:8080`.

### Swagger UI & OpenAPI Specification
- **Interactive Swagger UI**: Open [http://localhost:8080/swagger](http://localhost:8080/swagger) or [http://localhost:8080/docs](http://localhost:8080/docs) in your browser.
- **OpenAPI 3.0 Specification**: Available at [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json) or in the file [openapi.json](file:///d:/IT-projects/go-gin-todo/openapi.json).

### Run Integration Tests
```bash
go test -v ./tests/...
```

---

## API Endpoints Reference

Base prefixes:
- `/api/v1` (explicit version)
- `/api` (default version, mapped to `v1`)

Protected routes require:
```http
Authorization: Bearer <your_jwt_token>
```

### 1. Authentication

#### Register
- `POST /api/v1/register` (or `/api/register`)
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "secretpassword"
  }
  ```
- **Response** (`201 Created`):
  ```json
  {
    "message": "User registered successfully",
    "data": {
      "id": 1,
      "email": "user@example.com",
      "createdAt": "2026-11-17T21:00:00Z"
    }
  }
  ```

#### Login
- `POST /api/v1/login` (or `/api/login`)
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "secretpassword"
  }
  ```
- **Response** (`200 OK`):
  ```json
  {
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsIn...",
    "data": {
      "id": 1,
      "email": "user@example.com",
      "createdAt": "2026-11-17T21:00:00Z"
    }
  }
  ```

### Pagination Format

Both `GET /api/v1/todolist` and `GET /api/v1/todo` accept standard pagination query parameters:
- `page`: Page number (1-based, default: `1`)
- `pageSize`: Items per page (default: `10`, accepts aliases `limit` or `per_page`)

#### Example Paginated Response:
```json
{
  "data": [
    {
      "id": 1,
      "userId": 1,
      "title": "Work Tasks",
      "createdAt": "2026-11-17T21:00:00Z",
      "updatedAt": "2026-11-17T21:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 10,
    "totalItems": 42,
    "totalPages": 5,
    "hasNext": true,
    "hasPrev": false
  }
}
```

---

### 2. Todolist Endpoints (`/todolist`)

| Verb | Path | Description |
|---|---|---|
| `GET` | `/api/v1/todolist` | Bulk list for user with pagination (`?page=1`, `?pageSize=10`, `?search=...`) |
| `GET` | `/api/v1/todolist/:id` | Get single todolist by ID |
| `POST` | `/api/v1/todolist` | Create single `{"title": "..."}` or bulk `[{"title": "..."}, ...]` |
| `PATCH` | `/api/v1/todolist` | Update single `{"id": 1, "title": "..."}` or bulk `[{"id": 1, ...}]` |
| `PATCH` | `/api/v1/todolist/:id` | Update single by URL ID `{"title": "..."}` |
| `DELETE` | `/api/v1/todolist` | Delete strictly via body `[1, 2]` or `{"id": 1}` or `{"ids": [1, 2]}` |
| `DELETE` | `/api/v1/todolist/:id` | Delete single by URL ID |

---

### 3. Todo Endpoints (`/todo`)

| Verb | Path | Description |
|---|---|---|
| `GET` | `/api/v1/todo` | Bulk list for user with pagination (`?page=1`, `?pageSize=10`, `?check=true\|false`, `?search=...`) |
| `GET` | `/api/v1/todo/:id` | Get single todo by ID |
| `POST` | `/api/v1/todo` | Create single `{"message": "...", "check": false}` or bulk `[...]` |
| `PATCH` | `/api/v1/todo` | Update single `{"id": 1, "check": true}` or bulk `[{"id": 1, ...}]` |
| `PATCH` | `/api/v1/todo/:id` | Update single by URL ID `{"message": "...", "check": true}` |
| `DELETE` | `/api/v1/todo` | Delete strictly via body `[1, 2]` or `{"id": 1}` or `{"ids": [1, 2]}` |
| `DELETE` | `/api/v1/todo/:id` | Delete single by URL ID |

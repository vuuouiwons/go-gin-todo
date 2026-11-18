# Go Gin Todo Backend Server (Pragmatic Layered Architecture)

A pragmatic, low-overhead RESTful backend API built with Go and the [Gin Web Framework](https://github.com/gin-gonic/gin), structured according to the **Pragmatic Layered Architecture** defined in [AGENTS.md](file:///d:/IT-projects/go-gin-todo/AGENTS.md).

It features JWT authentication, bcrypt password hashing, pure-Go SQLite persistence with GORM (no CGO required), URI versioning (`v1` as default), lightweight CQRS (`Commands` & `Queries`), and full single/bulk CRUD capabilities for Todolists and Todos.

---

## Pragmatic Layered Architecture

Dependencies flow strictly **inward** toward the `Application` core (`API` → `Application` ← `Infrastructure`), with no `API` ↔ `Infrastructure` cross-imports:

```text
go-gin-todo/
├── docs/
│   ├── docs.go                                   # Embedded OpenAPI specification (//go:embed)
│   └── openapi.json                              # OpenAPI 3.0 specification
├── internal/
│   ├── app/                                      # 1. Application Layer (Core)
│   │   ├── models.go                             # Core models (User, Todolist, Todo, Pagination) & errors
│   │   ├── interfaces.go                         # Repository, Transactor (+ RunInTx), Queryer, Hasher, Token
│   │   ├── auth_command.go                       # Register & Login commands
│   │   ├── todolist_command.go                   # Todolist write commands & RunInTx orchestration
│   │   ├── todolist_query.go                     # Todolist reads (direct paginated Queryer + GetByID)
│   │   ├── todo_command.go                       # Todo write commands & RunInTx orchestration
│   │   ├── todo_query.go                         # Todo reads (direct paginated Queryer + GetByID)
│   │   └── commands_test.go                      # Fast in-memory Unit Tests for Commands
│   │
│   ├── infra/                                    # 2. Infrastructure Layer (Thin ACL)
│   │   ├── database.go                           # SQLite connection, migrations, Transactor & Queryer
│   │   ├── user_repo.go                          # GORM UserRepository (reuses app.User)
│   │   ├── todolist_repo.go                      # GORM TodolistRepository (reuses app.Todolist)
│   │   ├── todo_repo.go                          # GORM TodoRepository (reuses app.Todo)
│   │   └── security.go                           # BcryptHasher & JWTTokenManager
│   │
│   ├── api/                                      # 3. API Layer (HTTP Delivery & Security Boundary)
│   │   ├── dto.go                                # Request/Response DTOs & explicit mappers
│   │   ├── middleware.go                         # AuthMiddleware & CORSMiddleware
│   │   ├── helpers.go                            # Pagination, ID/body parsing, error mapping
│   │   ├── auth_handler.go                       # Auth HTTP handlers
│   │   ├── todolist_handler.go                   # Todolist HTTP handlers
│   │   ├── todo_handler.go                       # Todo HTTP handlers
│   │   └── router.go                             # Route definitions & URI versioning
│   │
│   └── config/                                   # Environment & .env configuration loader
│       └── config.go
├── tests/
│   ├── integration_test.go                       # Isolated DB tests for Queries, Repos, & Transactions
│   └── api_test.go                               # HTTP E2E tests across all endpoints
├── .dockerignore
├── .env.example
├── .gitignore
├── Dockerfile
├── go.mod
├── go.sum
├── main.go                                       # Composition root & server entry point
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

The server listens on `http://localhost:8080`.

### Run with Docker
```bash
docker build -t go-gin-todo .
docker run -d --name go-gin-todo -p 8080:8080 -e JWT_SECRET=change-me -v todo_data:/data go-gin-todo
```

The container persists SQLite data in `/data/todo.db`, runs as non-root `appuser`, and exposes port `8080`.

### Swagger UI & OpenAPI Specification
- **Interactive Swagger UI**: Open [http://localhost:8080/docs](http://localhost:8080/docs) in your browser.
- **OpenAPI 3.0 Specification**: Available at [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json) or in the file [docs/openapi.json](file:///d:/IT-projects/go-gin-todo/docs/openapi.json).

### Run Tests (Unit, Integration, and E2E)
```bash
go test -v ./...
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
- `limit`: Items per page (default: `10`)

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
    "limit": 10,
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
| `GET` | `/api/v1/todolist` | Bulk list for user with pagination (`?page=1`, `?limit=10`, `?search=...`) |
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
| `GET` | `/api/v1/todo` | Bulk list for user with pagination (`?page=1`, `?limit=10`, `?check=true\|false`, `?search=...`) |
| `GET` | `/api/v1/todo/:id` | Get single todo by ID |
| `POST` | `/api/v1/todo` | Create single `{"message": "...", "check": false}` or bulk `[...]` |
| `PATCH` | `/api/v1/todo` | Update single `{"id": 1, "check": true}` or bulk `[{"id": 1, ...}]` |
| `PATCH` | `/api/v1/todo/:id` | Update single by URL ID `{"message": "...", "check": true}` |
| `DELETE` | `/api/v1/todo` | Delete strictly via body `[1, 2]` or `{"id": 1}` or `{"ids": [1, 2]}` |
| `DELETE` | `/api/v1/todo/:id` | Delete single by URL ID |

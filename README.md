# Go Gin Todo Backend Server (Pragmatic 3-Layer Architecture)

A pragmatic, low-overhead RESTful backend API built with Go and the [Gin Web Framework](https://github.com/gin-gonic/gin), structured according to the **Pragmatic 3-Layer Architecture** defined in [AGENTS.md](file:///d:/IT-projects/go-gin-todo/AGENTS.md).

It features JWT authentication, bcrypt password hashing, pure-Go SQLite persistence with GORM (no CGO required), URI versioning (`v1` as default), direct ORM data access, idiomatic GORM transaction handling, and full single/bulk CRUD capabilities for Todolists and Todos.

---

## Pragmatic 3-Layer Architecture

**Data & Dependency Flow:**
`Inbound Interface (HTTP)` ➔ `Controllers / Handlers (src/routes)` ➔ `Services (src/services)` ➔ `Data Access (src/db)`

```text
go-gin-todo/
├── docs/
│   ├── docs.go                                   # Embedded OpenAPI specification (//go:embed)
│   └── openapi.json                              # OpenAPI 3.0 specification
├── src/
│   ├── routes/                                   # 1. Routes & Handlers (Transport & Response Shaping)
│   │   ├── dto.go                                # Request & Response DTOs (Boundary Security)
│   │   ├── helpers.go                            # Pagination, ID/body parsing, error mapping
│   │   ├── middleware.go                         # AuthMiddleware & CORSMiddleware
│   │   ├── handlers_auth.go                      # Auth HTTP handlers (Register & Login)
│   │   ├── handlers_todolist.go                  # Todolist HTTP handlers (CRUD & bulk)
│   │   ├── handlers_todo.go                      # Todo HTTP handlers (CRUD & bulk)
│   │   └── router.go                             # Route definitions & URI versioning
│   │
│   ├── services/                                 # 2. Services (Business Logic & Transactions)
│   │   ├── auth_service.go                       # Registration, password hashing, JWT issuing
│   │   ├── todolist_service.go                   # Todolist business rules & GORM transactions
│   │   ├── todo_service.go                       # Todo business rules & GORM transactions
│   │   └── services_test.go                      # Unit tests for business logic & transaction rollbacks
│   │
│   ├── db/                                       # 3. Data Access (GORM Connection & Models)
│   │   ├── connection.go                         # SQLite connection & auto-migrations
│   │   └── models.go                             # GORM schema models (User, Todolist, Todo, Pagination)
│   │
│   └── utils/                                    # Stateless Helpers
│       ├── auth.go                               # Bcrypt hashing & JWT token manager
│       ├── errors.go                             # Shared application domain errors
│       ├── pagination.go                         # Pagination helpers & calculation
│       └── config.go                             # Environment configuration loader
├── tests/
│   ├── integration_test.go                       # Isolated DB tests for queries & transactions
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

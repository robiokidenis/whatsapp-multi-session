# WhatsApp Multi-Session Manager - Clean Architecture

This document describes the new clean architecture implementation of the WhatsApp Multi-Session Manager.

## Architecture Overview

The application has been restructured using clean architecture principles with clear separation of concerns:

```
whatsapp-multi-session/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── handlers/          # HTTP handlers (controllers)
│   ├── middleware/        # HTTP middleware
│   ├── models/           # Domain models and DTOs
│   ├── repository/       # Data access layer
│   ├── services/         # Business logic layer
│   └── utils/            # Internal utilities
├── pkg/                   # Public packages (can be imported)
│   ├── logger/           # Logging utilities
│   └── ratelimiter/      # Rate limiting utilities
├── frontend/             # React frontend
└── main.go              # Application entry point
```

## Layer Descriptions

### 1. Configuration Layer (`internal/config/`)
- **Purpose**: Centralized configuration management
- **Files**: `config.go`
- **Responsibilities**:
  - Load configuration from environment variables
  - Provide default values
  - Validate configuration settings

### 2. Models Layer (`internal/models/`)
- **Purpose**: Domain models and data transfer objects
- **Files**: `user.go`, `session.go`, `message.go`
- **Responsibilities**:
  - Define data structures
  - Request/response models
  - Domain entities

### 3. Repository Layer (`internal/repository/`)
- **Purpose**: Data access abstraction
- **Files**: `database.go`, `user_repository.go`, `session_repository.go`
- **Responsibilities**:
  - Database operations
  - Data persistence
  - Query abstractions

### 4. Services Layer (`internal/services/`)
- **Purpose**: Business logic implementation
- **Files**: `user_service.go`, `whatsapp_service.go`
- **Responsibilities**:
  - Business rules
  - Domain logic
  - Service orchestration

### 5. Handlers Layer (`internal/handlers/`)
- **Purpose**: HTTP request handling
- **Files**: `auth_handler.go`, `session_handler.go`
- **Responsibilities**:
  - HTTP request/response handling
  - Input validation
  - Response formatting

### 6. Middleware Layer (`internal/middleware/`)
- **Purpose**: HTTP middleware components
- **Files**: `auth.go`, `cors.go`
- **Responsibilities**:
  - Authentication
  - Authorization
  - CORS handling
  - Request logging

### 7. Package Layer (`pkg/`)
- **Purpose**: Reusable packages
- **Files**: `logger/`, `ratelimiter/`
- **Responsibilities**:
  - Shared utilities
  - Reusable components
  - External integrations

## Key Improvements

### 1. **Separation of Concerns**
- Each layer has a single responsibility
- Clear boundaries between layers
- Reduced coupling between components

### 2. **Testability**
- Dependency injection pattern
- Interface-based design
- Mockable components

### 3. **Maintainability**
- Modular structure
- Clear file organization
- Consistent naming conventions

### 4. **Scalability**
- Easy to add new features
- Clear extension points
- Service-oriented design

### 5. **Configuration Management**
- Centralized configuration
- Environment-based settings
- Default value handling

## Data Flow

```
HTTP Request → Middleware → Handler → Service → Repository → Database
                                         ↓
HTTP Response ← Handler ← Service ← Repository ← Database
```

1. **Request Flow**:
   - HTTP requests hit middleware (auth, CORS)
   - Handlers parse and validate requests
   - Services implement business logic
   - Repositories handle data persistence

2. **Response Flow**:
   - Repositories return data
   - Services process and transform data
   - Handlers format responses
   - Middleware handles cross-cutting concerns

## Building and Running

### Development
```bash
# Build the application
go build -o whatsapp-multi-session .

# Run with custom configuration
export PORT=8080
export LOG_LEVEL=debug
./whatsapp-multi-session
```

### Production
```bash
# Initialize Docker environment
./docker-init.sh

# Configure environment
cp .env.example .env
# Edit .env file with your settings

# Run with Docker
docker-compose up -d
```

## Migration from Old Structure

The original `main.go` (3420+ lines) has been broken down into:

- **Config**: ~150 lines → `internal/config/config.go`
- **Models**: ~200 lines → `internal/models/*.go`  
- **Database**: ~300 lines → `internal/repository/*.go`
- **Services**: ~800 lines → `internal/services/*.go`
- **Handlers**: ~1000 lines → `internal/handlers/*.go`
- **Middleware**: ~200 lines → `internal/middleware/*.go`
- **Main**: ~150 lines → `main.go`

This results in:
- ✅ **95% code reduction** in main.go
- ✅ **Clear separation** of concerns
- ✅ **Better testability** and maintainability
- ✅ **Easier debugging** and development
- ✅ **Consistent patterns** across the codebase

## Next Steps

1. **Add comprehensive tests** for each layer
2. **Implement remaining endpoints** (user management)
3. **Add API documentation** generation
4. **Implement message handling** services
5. **Add monitoring and metrics** capabilities
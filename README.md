# Platform Service

A Go-based backend platform service featuring user, organization, and role management with a multi-layer persistence model using GORM, PostgreSQL, and a configurable cache layer.

## Features

- GORM-based PostgreSQL database integration
- Configurable cache layer with Redis and Elasticsearch implementations
- RESTful API with Gin framework
- Comprehensive model validation
- Multi-layer persistence policy: validate first, cache, then store in database
- Support for user, organization, and role models
- Authentication with JWT
- Multiple OAuth providers support (Google, Microsoft, GitHub, WeChat)

## Project Structure

```
platform/
├── api/
│   └── v1/               # API v1 handlers
├── cmd/
│   └── api/              # API server entry point
├── config/               # Configuration files
├── docs/                 # Documentation
├── internal/
│   ├── cache/            # Cache layer implementations
│   ├── model/            # Data models
│   └── service/          # Business logic and repository
├── pkg/
│   ├── config/           # Configuration utilities
│   └── middleware/       # HTTP middleware
├── scripts/              # Utility scripts
└── main.go               # Main application entry point
```

## Prerequisites

- Go 1.24.2 or later
- PostgreSQL 12 or later
- Redis or Elasticsearch (depending on cache configuration)

## Setup

1. Clone the repository
2. Copy the sample configuration file:
   ```bash
   cp config/config.example.yml config/config.yml
   ```
3. Modify the configuration in `config/config.yml` to match your environment
4. Start the dependencies (PostgreSQL and Redis or Elasticsearch)
5. Run the API server:
   ```bash
   go run cmd/api/main.go
   ```

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login with username/email and password
- `POST /api/v1/auth/refresh-token` - Refresh JWT token
- `GET /api/v1/auth/google` - Initiate Google OAuth login
- `GET /api/v1/auth/google/callback` - Google OAuth callback
- `GET /api/v1/auth/microsoft` - Initiate Microsoft OAuth login
- `GET /api/v1/auth/microsoft/callback` - Microsoft OAuth callback
- `GET /api/v1/auth/github` - Initiate GitHub OAuth login
- `GET /api/v1/auth/github/callback` - GitHub OAuth callback
- `GET /api/v1/auth/wechat` - Initiate WeChat OAuth login
- `GET /api/v1/auth/wechat/callback` - WeChat OAuth callback

### Users

- `GET /api/v1/users` - List users with pagination
- `GET /api/v1/users/:id` - Get user by ID
- `GET /api/v1/users/username/:username` - Get user by username
- `GET /api/v1/users/email/:email` - Get user by email
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Organizations

- `GET /api/v1/organizations` - List organizations with pagination
- `GET /api/v1/organizations/:id` - Get organization by ID
- `GET /api/v1/organizations/name/:name` - Get organization by name
- `POST /api/v1/organizations` - Create new organization
- `PUT /api/v1/organizations/:id` - Update organization
- `DELETE /api/v1/organizations/:id` - Delete organization

### Roles

- `GET /api/v1/roles` - List roles with pagination
- `GET /api/v1/roles/:id` - Get role by ID
- `GET /api/v1/roles/name/:name` - Get role by name
- `POST /api/v1/roles` - Create new role
- `PUT /api/v1/roles/:id` - Update role
- `DELETE /api/v1/roles/:id` - Delete role

## Persistence Policy

The service implements a multi-layer persistence strategy:

1. **Validation**: All data is validated before any persistence operations
2. **Cache**: After successful validation and database operations, data is cached
3. **Database**: The database is the ultimate source of truth

This approach provides:
- Fast response times through caching
- Data integrity through validation
- Reliability through persistent database storage

## Cache Configuration

The service supports two caching backends:

### Redis

```yaml
cache:
  type: redis
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

### Elasticsearch

```yaml
cache:
  type: elasticsearch
elasticsearch:
  addresses:
    - http://localhost:9200
  username: ""
  password: ""
  index_name: platform_cache
```

## Running Tests

```bash
go test -v ./...
```

## License

[Apache License 2.0](LICENSE)
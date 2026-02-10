# Chat SDK API Service

A multi-tenant, API-first chat service built with Go, Gin, WebSockets, and PostgreSQL. This service allows third-party applications to integrate chat functionality by consuming REST APIs and maintaining WebSocket connections for real-time message delivery.

## Features

- **Multi-Tenant Architecture**: Support multiple applications with complete data isolation
- **RESTful APIs**: Complete REST API for sending messages, managing rooms, users, and reactions
- **Real-Time Communication**: WebSocket connections for instant message delivery
- **Message Reactions**: Support for emoji reactions on messages
- **Typing Indicators**: Real-time typing status updates
- **Message History**: Paginated message retrieval with cursor-based pagination
- **Read Receipts**: Track message read status
- **File Attachments**: Support for file uploads and attachments

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architecture documentation.

See [HOW_IT_WORKS.md](./HOW_IT_WORKS.md) for detailed explanation of how the system works.

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15+
- **WebSocket**: Gorilla WebSocket
- **JWT**: golang-jwt/jwt
- **Docker**: For local development

## Prerequisites


- Go 1.21 or higher
- PostgreSQL 15+ (or use Docker Compose)
- Docker and Docker Compose (for local development)
- Make (optional, for using Makefile commands)

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd chat-app
```

### 2. Set Up Environment Variables

Copy the example environment file and configure:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=chat_app_user
DB_PASSWORD=chat_app_password
DB_NAME=chat_app_db

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Server
SERVER_PORT=8080
ENVIRONMENT=development
```

### 3. Start Database with Docker

```bash
make docker-up
# Or manually:
docker-compose up -d
```

This will start:
- PostgreSQL on port 5432
- Redis on port 6379 (optional, for scaling)

### 4. Run Database Migrations

```bash
make migrate
# Or manually:
psql -h localhost -U chat_app_user -d chat_app_db -f internal/database/migrations/001_initial_schema.sql
```

### 5. Install Dependencies

```bash
make deps
# Or manually:
go mod download
go mod tidy
```

### 6. Run the Application

```bash
make run
# Or manually:
go run ./cmd/server
```

The server will start on `http://localhost:8080`

### 7. Test the Health Endpoint

```bash
curl http://localhost:8080/health
```

## Project Structure

```
chat-app/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   ├── postgres.go          # Database connection
│   │   └── migrations/          # Database migrations
│   ├── handler/
│   │   ├── rest/                # REST API handlers
│   │   └── websocket/           # WebSocket handlers
│   ├── middleware/              # HTTP middleware
│   ├── models/                  # Data models (TODO)
│   ├── repository/              # Data access layer (TODO)
│   ├── service/                 # Business logic (TODO)
│   └── utils/                   # Utility functions
├── docker-compose.yml           # Docker services
├── Makefile                     # Build automation
├── go.mod                       # Go dependencies
├── ARCHITECTURE.md              # Architecture documentation
└── README.md                    # This file
```

## API Endpoints

### Health Checks

- `GET /health` - Basic health check
- `GET /health/ready` - Readiness check (database connection)
- `GET /health/live` - Liveness check

### Authentication (TODO: Implementation needed)

- `POST /api/v1/auth/register` - Register application
- `POST /api/v1/auth/login` - Login and get JWT token
- `POST /api/v1/auth/refresh` - Refresh access token

### Applications (TODO: Implementation needed)

- `POST /api/v1/applications` - Create application
- `GET /api/v1/applications` - List applications
- `GET /api/v1/applications/{id}` - Get application
- `PUT /api/v1/applications/{id}` - Update application
- `DELETE /api/v1/applications/{id}` - Delete application

### Users (TODO: Implementation needed)

- `POST /api/v1/applications/{app_id}/users` - Create user
- `GET /api/v1/applications/{app_id}/users` - List users
- `GET /api/v1/applications/{app_id}/users/{id}` - Get user
- `PUT /api/v1/applications/{app_id}/users/{id}` - Update user
- `DELETE /api/v1/applications/{app_id}/users/{id}` - Delete user

### Rooms (TODO: Implementation needed)

- `POST /api/v1/applications/{app_id}/rooms` - Create room
- `GET /api/v1/applications/{app_id}/rooms` - List rooms
- `GET /api/v1/applications/{app_id}/rooms/{id}` - Get room
- `PUT /api/v1/applications/{app_id}/rooms/{id}` - Update room
- `DELETE /api/v1/applications/{app_id}/rooms/{id}` - Delete room
- `POST /api/v1/applications/{app_id}/rooms/{id}/members` - Add member
- `DELETE /api/v1/applications/{app_id}/rooms/{id}/members/{user_id}` - Remove member
- `GET /api/v1/applications/{app_id}/rooms/{id}/members` - List members

### Messages (TODO: Implementation needed)

- `POST /api/v1/applications/{app_id}/rooms/{room_id}/messages` - Send message
- `GET /api/v1/applications/{app_id}/rooms/{room_id}/messages` - Get message history
- `GET /api/v1/applications/{app_id}/messages/{message_id}` - Get message
- `PUT /api/v1/applications/{app_id}/messages/{message_id}` - Edit message
- `DELETE /api/v1/applications/{app_id}/messages/{message_id}` - Delete message

### Reactions (TODO: Implementation needed)

- `POST /api/v1/applications/{app_id}/messages/{message_id}/reactions` - Add reaction
- `DELETE /api/v1/applications/{app_id}/messages/{message_id}/reactions/{reaction_id}` - Remove reaction
- `GET /api/v1/applications/{app_id}/messages/{message_id}/reactions` - List reactions

### Typing Indicators (TODO: Implementation needed)

- `POST /api/v1/applications/{app_id}/rooms/{room_id}/typing` - Send typing indicator

### WebSocket

- `GET /api/v1/ws?token={jwt_token}&app_id={app_id}` - WebSocket connection endpoint

## Development

### Running Tests

```bash
make test
```

### Running with Coverage

```bash
make test-coverage
```

### Code Formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

### Building Binary

```bash
make build
```

The binary will be in `bin/chat-app`

## Docker Commands

### Start Services

```bash
make docker-up
```

### Stop Services

```bash
make docker-down
```

### View Logs

```bash
make docker-logs
```

## Next Steps

The starter repository includes:

✅ Project structure  
✅ Configuration management  
✅ Database connection setup  
✅ Migration files  
✅ Middleware (Auth, CORS, Logger, Recovery)  
✅ Handler skeletons  
✅ WebSocket hub structure  
✅ Utility functions (JWT, Errors, Validator)  
✅ Docker Compose setup  
✅ Makefile for common tasks  

**TODO: Implement the following:**

1. **Models** - Define Go structs for all database entities
2. **Repositories** - Implement data access layer with proper queries
3. **Services** - Implement business logic layer
4. **Handlers** - Complete all REST and WebSocket handler implementations
5. **Authentication** - Implement registration and login flows
6. **Message Delivery** - Complete WebSocket message broadcasting
7. **Testing** - Add comprehensive unit and integration tests
8. **Documentation** - Add API documentation (OpenAPI/Swagger)

## Contributing

1. Follow the Go code style guidelines
2. Ensure all tests pass
3. Update documentation as needed
4. Follow the architecture patterns defined in ARCHITECTURE.md

## License

[Your License Here]

## Support

For issues and questions, please open an issue on the repository.


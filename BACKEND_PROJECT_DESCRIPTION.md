# Chat Application Backend - Project Description

## Overview

A production-ready, multi-tenant chat service backend built with Go and Gin framework. This API-first service enables third-party applications to integrate real-time chat functionality through comprehensive REST APIs and WebSocket connections. The system supports complete data isolation between applications, real-time message delivery, and scalable architecture designed for high-performance chat applications.

## Key Features

**Multi-Tenant Architecture**: Complete application-level isolation with separate data spaces for each integrated application. Each application has its own users, rooms, and messages with enforced access controls.

**Real-Time Communication**: WebSocket-based hub architecture for instant message broadcasting, typing indicators, and live updates. Supports room-based subscriptions and efficient message routing to connected clients.

**Comprehensive REST API**: Full CRUD operations for applications, users, rooms, messages, reactions, and typing indicators. RESTful design with proper HTTP status codes, error handling, and standardized response formats.

**Authentication & Security**: JWT-based authentication with access and refresh token support. Role-based access control (admin, member, owner) for room management. CORS configuration for cross-origin requests. Input validation and SQL injection protection.

**Message Management**: Support for text messages, message editing, soft deletion, message reactions (emoji), reply-to functionality, and message metadata. Paginated message retrieval with offset-based pagination.

**Room Management**: Group and direct message room types. Member management with roles (admin, member, owner). Room metadata and descriptions. Last read tracking for unread message indicators.

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin (high-performance HTTP web framework)
- **Database**: PostgreSQL 15+ with optimized indexes and foreign key constraints
- **WebSocket**: Gorilla WebSocket for real-time bidirectional communication
- **Authentication**: golang-jwt/jwt for JWT token generation and validation
- **Logging**: Structured logging with configurable levels (debug, info, warn, error)
- **Configuration**: Environment-based configuration with validation
- **Deployment**: Docker support, graceful shutdown, health check endpoints

## Architecture

**Layered Architecture**: 
- **Handlers Layer**: REST API handlers and WebSocket handlers for request processing
- **Middleware Layer**: Authentication, CORS, logging, and recovery middleware
- **Database Layer**: PostgreSQL connection pooling with connection lifecycle management
- **WebSocket Hub**: Centralized hub for managing connections, room subscriptions, and message broadcasting

**Request Flow**: 
1. HTTP request → Middleware (CORS, Logger, Auth) → Handler → Database → Response
2. WebSocket: Connection upgrade → Authentication → Hub registration → Room subscription → Message broadcasting

**Database Design**: 
- Normalized schema with proper foreign key relationships
- Indexed queries for optimal performance (room_id, user_id, created_at)
- Soft deletion pattern (deleted_at timestamps)
- JSONB metadata fields for flexible data storage
- Automatic timestamp management via database triggers

## API Endpoints

**Authentication** (`/api/v1/auth`):
- `POST /register` - Application registration
- `POST /login` - User login with JWT token generation
- `POST /refresh` - Access token refresh

**Applications** (`/api/v1/applications`):
- `POST` - Create application
- `GET` - Get application details
- `PATCH` - Update application
- `DELETE` - Delete application

**Users** (`/api/v1/users`):
- `POST` - Create user
- `GET` - Get user details
- `PATCH` - Update user
- `DELETE` - Delete user
- `GET /all` - List all users

**Rooms** (`/api/v1/rooms`):
- `POST` - Create room
- `GET` - List rooms
- `PATCH` - Update room
- `DELETE` - Delete room
- `POST /members` - Add room member
- `DELETE /members` - Remove room member
- `GET /members` - List room members

**Messages** (`/api/v1/messages`):
- `POST` - Send message
- `GET` - List messages (paginated with limit/offset)
- `PATCH` - Update message
- `DELETE` - Delete message

**Reactions** (`/api/v1/reactions`):
- `POST` - Add reaction to message
- `DELETE` - Remove reaction
- `GET` - List message reactions

**Typing Indicators** (`/api/v1/typing`):
- `POST` - Send typing indicator

**WebSocket** (`/api/v1/ws`):
- `GET /ws?token={jwt_token}` - WebSocket connection endpoint
- Supports room subscription/unsubscription
- Real-time message broadcasting
- Typing indicator broadcasting

**Health Checks**:
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness probe (database connectivity)
- `GET /health/live` - Liveness probe

## Database Schema

**Core Tables**:
- `applications` - Multi-tenant application registry
- `users` - User accounts with application association
- `rooms` - Chat rooms (group or direct message)
- `room_members` - Room membership with roles
- `messages` - Chat messages with soft deletion
- `message_reactions` - Emoji reactions on messages
- `message_reads` - Read receipt tracking
- `typing_indicators` - Real-time typing status
- `files` - File attachments (future support)

**Key Relationships**:
- Applications → Users (one-to-many)
- Applications → Rooms (one-to-many)
- Rooms → Messages (one-to-many)
- Messages → Reactions (one-to-many)
- Rooms ↔ Users (many-to-many via room_members)

## Security Features

- **JWT Authentication**: Secure token-based authentication with configurable expiry
- **Role-Based Access Control**: Room-level permissions (admin, member, owner)
- **Input Validation**: Request validation with detailed error messages
- **SQL Injection Protection**: Parameterized queries throughout
- **CORS Configuration**: Configurable allowed origins, methods, and headers
- **Error Handling**: Standardized error responses with proper HTTP status codes
- **Soft Deletion**: Data retention with soft delete pattern

## Performance & Scalability

- **Connection Pooling**: Configurable database connection pool (max connections, idle connections, connection lifetime)
- **Indexed Queries**: Optimized database indexes for common query patterns
- **WebSocket Hub**: Efficient message broadcasting with room-based subscriptions
- **Graceful Shutdown**: Proper cleanup of connections and resources
- **Health Monitoring**: Readiness and liveness probes for container orchestration
- **Structured Logging**: JSON-formatted logs for production monitoring

## Deployment

**Configuration**: Environment-based configuration via `.env` files or environment variables. Supports development and production modes.

**Docker Support**: Docker Compose setup for local development with PostgreSQL and optional Redis.

**Production Ready**: 
- Graceful shutdown handling
- Health check endpoints for load balancers
- Configurable timeouts and connection limits
- Structured logging for log aggregation
- CORS configuration for frontend integration

**Scalability Considerations**: 
- Stateless API design enables horizontal scaling
- WebSocket hub can be extended with Redis pub/sub for multi-instance deployments
- Database connection pooling handles concurrent requests efficiently
- Optional Kafka integration for event streaming (prepared but not active)

## Development

**Project Structure**: Clean architecture with separation of concerns (handlers, middleware, models, database, utils).

**Code Quality**: 
- Error handling with context-rich error messages
- Comprehensive logging for debugging
- Input validation and sanitization
- Type-safe database operations

**Testing**: Prepared structure for unit and integration tests.

---

*This backend service provides a complete foundation for building chat-enabled applications with enterprise-grade features, security, and scalability.*


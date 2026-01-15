# Environment Setup Guide

## Creating .env File

Create a `.env` file in the root directory with the following configuration:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=chat_app_user
DB_PASSWORD=t
DB_NAME=chat_app_db
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=300s

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
CORS_ALLOW_CREDENTIALS=true

# Redis Configuration (for scaling - optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_ENABLED=false

# File Storage Configuration
FILE_STORAGE_PATH=./uploads
MAX_FILE_SIZE=10485760
ALLOWED_FILE_TYPES=image/jpeg,image/png,image/gif,application/pdf,text/plain

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

## Important Notes

1. **JWT_SECRET**: In production, use a strong, randomly generated secret key (at least 32 characters)
2. **DB_PASSWORD**: Change the default password in production
3. **CORS_ALLOWED_ORIGINS**: Add your frontend application URLs (comma-separated)
4. **ENVIRONMENT**: Set to `production` when deploying

## Quick Setup Command

```bash
cat > .env << 'EOF'
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=chat_app_user
DB_PASSWORD=chat_app_password
DB_NAME=chat_app_db
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=300s
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
CORS_ALLOW_CREDENTIALS=true
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_ENABLED=false
FILE_STORAGE_PATH=./uploads
MAX_FILE_SIZE=10485760
ALLOWED_FILE_TYPES=image/jpeg,image/png,image/gif,application/pdf,text/plain
LOG_LEVEL=info
LOG_FORMAT=json
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
EOF
```


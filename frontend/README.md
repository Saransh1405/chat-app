# Chat App - Minimal Frontend

A minimal frontend implementation that integrates with all backend APIs of the chat application.

## Overview

This is a complete, minimal frontend that demonstrates integration with all backend endpoints. It includes:

- Full authentication flow (register/login)
- Real-time chat with WebSocket support
- Room management
- Message sending and receiving
- Reactions on messages
- Typing indicators
- All CRUD operations for applications, users, and rooms

**Note**: The backend WebSocket handler has been updated to accept authentication tokens via query parameters (since browsers can't set custom headers for WebSocket connections). This change is in `internal/handler/websocket/handler.go` and `cmd/server/main.go`.

## Features

- **Authentication**: Register and Login
- **Applications**: Create, Get, List, Update, Delete
- **Users**: Create, Get, Update, Delete
- **Rooms**: Create, Get, List, Update, Delete, Add/Remove Members, List Members
- **Messages**: Create, Get, List, Update, Delete
- **Reactions**: Create, Delete, List
- **Typing Indicators**: Real-time typing indicators
- **WebSocket**: Real-time messaging, reactions, typing indicators, room updates

## Setup

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The app will be available at `http://localhost:3000`

## Configuration

The frontend is configured to connect to the backend at `http://localhost:8080` by default. You can change this by setting environment variables:

- `VITE_API_URL`: Backend API URL (default: `http://localhost:8080/api/v1`)
- `VITE_WS_URL`: WebSocket URL (default: `ws://localhost:8080/api/v1/ws`)

## Usage

1. **Register/Login**: Use the login page to register a new user or login with an existing email
2. **Create Room**: Click "Create Room" button in the sidebar
3. **Select Room**: Click on any room in the sidebar to open it
4. **Send Messages**: Type a message and press Enter or click Send
5. **Add Reactions**: Click "+ React" on any message to add a reaction
6. **View Typing Indicators**: See when users are typing in real-time

## API Integration

All backend APIs are integrated:

- `/api/v1/auth/*` - Authentication endpoints
- `/api/v1/applications/*` - Application management
- `/api/v1/users/*` - User management
- `/api/v1/rooms/*` - Room management
- `/api/v1/messages/*` - Message management
- `/api/v1/reactions/*` - Reaction management
- `/api/v1/typing/*` - Typing indicators
- `/api/v1/ws` - WebSocket connection for real-time features

## Project Structure

```
minimal-frontend/
├── src/
│   ├── api/          # API client modules
│   ├── components/   # React components
│   ├── types/         # TypeScript type definitions
│   ├── App.tsx        # Main app component
│   └── main.tsx       # Entry point
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```


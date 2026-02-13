

# Aura Chat — Premium Glassmorphic Chat Application

## Overview
A stunning, high-end chat application with glassmorphic design, dark/light mode, and full real-time messaging features. Built to connect to your external REST + WebSocket backend.

---

## 1. Design System & Theming
- **Typography**: Import and apply the "Inter" font from Google Fonts for a clean, modern look
- **Glassmorphism**: Backdrop-blur effects, semi-transparent panels, and soft glowing shadows throughout the UI
- **Color palette**: Vibrant gradient accents (purple-to-blue, orange-to-pink) for avatars and active states, with muted neutral tones for backgrounds
- **Dark / Light mode toggle**: Full theme support with smooth transitions, matching the mockup's dual-mode aesthetic
- **Custom CSS variables**: Extended design tokens for glass effects, gradient borders, and glow shadows

---

## 2. Authentication Pages
- **Login page**: Email + password form with glassmorphic card styling, gradient accents, and JWT token handling
- **Signup page**: Registration form with the same premium aesthetic
- **Auth flow**: Store JWT in localStorage, attach as Bearer token to all API requests, redirect authenticated users to the chat
- **Protected routes**: Unauthenticated users are redirected to login

---

## 3. App Layout
- **Sidebar** (left panel): Glassmorphic panel with backdrop-blur, listing chat rooms with vibrant gradient avatars and active room highlighting
- **Main chat area** (right panel): Room header, scrollable message list, and floating input bar
- **Responsive**: Sidebar collapses on mobile with a hamburger toggle
- **Dark/light mode toggle** accessible from the sidebar or header

---

## 4. Sidebar — Room List
- List of available rooms with room name, member count, and a gradient avatar icon
- Active room is highlighted with a glowing gradient border/background
- **"Create Room"** button to open a room creation dialog (name + type)
- Visual distinction between rooms you're a member of vs. rooms you can join

---

## 5. Chat Area — Room Header
- Room name and member count displayed prominently
- **"Invite" button** (visible to admins): Opens an invite panel/dialog to add users by email or username
- **"Join" button**: Shown for non-members, allowing them to join the room

---

## 6. Message List
- Scrollable message area with auto-scroll to the latest message
- **Message bubbles**: Distinct styles for "my" messages (right-aligned, gradient background) vs "others" (left-aligned, glass card)
- Each message shows the sender's gradient avatar, username, timestamp, and content
- **Image attachments**: Inline image previews with a caption and filename, rendered inside the message bubble
- **Emoji reactions**: Row of emoji reaction badges displayed below each message, showing the emoji and count

---

## 7. Typing Indicators
- Positioned just above the message input area
- Shows "[User] is typing..." with an animated dot indicator (●●●)
- Driven by WebSocket presence events from your backend

---

## 8. Message Input Area
- Floating input bar at the bottom with glassmorphic styling
- Send button with gradient accent
- **Image attachment button**: File picker for uploading images (sends to your backend's upload endpoint)
- **Emoji picker button**: Opens a popover with common emoji to insert into the message
- Sends typing events via WebSocket as the user types

---

## 9. Service Layer (Placeholder Endpoints)
- **API services**: Clean, separated files for `authAPI`, `roomsAPI`, `messagesAPI`, `usersAPI` — each with placeholder REST endpoints you can swap with your real API docs
- **WebSocket client**: A `wsClient` module that manages connection, reconnection, and event dispatching for real-time messages, typing indicators, and presence
- **Data models**: TypeScript interfaces for `User` (id, username, email, avatar_url, application_id), `Room` (id, name, type, member_count), and `Message` (id, content, user_id, created_at, reactions[])
- **Auth interceptor**: Automatically attaches JWT Bearer token to all API requests

---

## 10. State Management
- React context for auth state (current user + JWT)
- React Query for fetching and caching rooms and messages
- Local state for active room selection, typing status, and UI toggles
- WebSocket events update the message list and typing indicators in real-time


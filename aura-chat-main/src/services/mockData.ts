import type { User, Room, Message } from "@/types/chat";

export const mockCurrentUser: User = {
  id: "user-1",
  username: "you",
  email: "you@aura.chat",
  avatar_url: null,
  application_id: "app-1",
  is_online: true,
};

export const mockUsers: User[] = [
  mockCurrentUser,
  { id: "user-2", username: "alex.design", email: "alex@aura.chat", avatar_url: null, application_id: "app-1", is_online: true },
  { id: "user-3", username: "sarah.dev", email: "sarah@aura.chat", avatar_url: null, application_id: "app-1", is_online: true },
  { id: "user-4", username: "mike.pm", email: "mike@aura.chat", avatar_url: null, application_id: "app-1", is_online: false },
  { id: "user-5", username: "emma.eng", email: "emma@aura.chat", avatar_url: null, application_id: "app-1", is_online: true },
];

export const mockRooms: Room[] = [
  { id: "room-1", name: "General", type: "public", member_count: 24, last_message: "Hey team! 🚀", last_message_at: "2025-02-07T10:30:00Z", unread_count: 3 },
  { id: "room-2", name: "Design System", type: "public", member_count: 8, last_message: "New tokens pushed", last_message_at: "2025-02-07T09:15:00Z", unread_count: 0 },
  { id: "room-3", name: "Backend Crew", type: "private", member_count: 5, last_message: "API v2 is live!", last_message_at: "2025-02-07T08:45:00Z", unread_count: 1 },
  { id: "room-4", name: "Random", type: "public", member_count: 42, last_message: "Check this meme 😂", last_message_at: "2025-02-06T23:00:00Z", unread_count: 12 },
  { id: "room-5", name: "Announcements", type: "public", member_count: 50, last_message: "Q1 roadmap update", last_message_at: "2025-02-06T18:30:00Z", unread_count: 0 },
];

const userMap = Object.fromEntries(mockUsers.map((u) => [u.id, u]));

export const mockMessages: Record<string, Message[]> = {
  "room-1": [
    {
      id: "msg-1", content: "Good morning team! Ready for the sprint review? 🚀", user_id: "user-2", room_id: "room-1", created_at: "2025-02-07T09:00:00Z",
      reactions: [{ emoji: "🚀", count: 3, users: ["user-1", "user-3", "user-5"] }],
      user: userMap["user-2"],
    },
    {
      id: "msg-2", content: "Absolutely! I've finished the glassmorphic sidebar component. Let me share a preview.", user_id: "user-3", room_id: "room-1", created_at: "2025-02-07T09:02:00Z",
      reactions: [{ emoji: "✨", count: 2, users: ["user-2", "user-1"] }, { emoji: "🔥", count: 1, users: ["user-5"] }],
      user: userMap["user-3"],
    },
    {
      id: "msg-3", content: "Looks incredible! The blur effects are so smooth.", user_id: "user-1", room_id: "room-1", created_at: "2025-02-07T09:05:00Z",
      reactions: [],
      user: userMap["user-1"],
    },
    {
      id: "msg-4", content: "I've also pushed the new color tokens. Check the design system channel for details.", user_id: "user-5", room_id: "room-1", created_at: "2025-02-07T09:10:00Z",
      reactions: [{ emoji: "👍", count: 4, users: ["user-1", "user-2", "user-3", "user-4"] }],
      user: userMap["user-5"],
    },
    {
      id: "msg-5", content: "The PM sync is at 2pm today. Everyone please prepare your updates.", user_id: "user-4", room_id: "room-1", created_at: "2025-02-07T10:00:00Z",
      reactions: [{ emoji: "✅", count: 3, users: ["user-1", "user-2", "user-3"] }],
      user: userMap["user-4"],
    },
    {
      id: "msg-6", content: "Hey team! 🚀", user_id: "user-2", room_id: "room-1", created_at: "2025-02-07T10:30:00Z",
      reactions: [],
      user: userMap["user-2"],
    },
  ],
  "room-2": [
    {
      id: "msg-7", content: "Just updated the gradient tokens — purple to pink looks amazing on dark mode.", user_id: "user-5", room_id: "room-2", created_at: "2025-02-07T08:30:00Z",
      reactions: [{ emoji: "💜", count: 2, users: ["user-1", "user-3"] }],
      user: userMap["user-5"],
    },
    {
      id: "msg-8", content: "New tokens pushed", user_id: "user-3", room_id: "room-2", created_at: "2025-02-07T09:15:00Z",
      reactions: [],
      user: userMap["user-3"],
    },
  ],
  "room-3": [
    {
      id: "msg-9", content: "WebSocket reconnection logic is solid now. Tested with 1000 concurrent connections.", user_id: "user-3", room_id: "room-3", created_at: "2025-02-07T08:00:00Z",
      reactions: [{ emoji: "💪", count: 2, users: ["user-1", "user-5"] }],
      user: userMap["user-3"],
    },
    {
      id: "msg-10", content: "API v2 is live!", user_id: "user-5", room_id: "room-3", created_at: "2025-02-07T08:45:00Z",
      reactions: [{ emoji: "🎉", count: 3, users: ["user-1", "user-2", "user-3"] }],
      user: userMap["user-5"],
    },
  ],
};

export const getUserById = (id: string): User | undefined => mockUsers.find((u) => u.id === id);

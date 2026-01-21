import { create } from 'zustand';
import type { User, Room, Message, TypingIndicator, Application } from '@/types/chat';

interface ChatState {
  // Auth
  currentUser: User | null;
  token: string | null;
  isAuthenticated: boolean;
  
  // Applications
  applications: Application[];
  currentApplication: Application | null;
  
  // Rooms
  rooms: Room[];
  currentRoom: Room | null;
  
  // Messages
  messages: Record<string, Message[]>;
  
  // Typing
  typingUsers: Record<string, TypingIndicator[]>;
  
  // Online users
  onlineUsers: Set<string>;
  
  // Actions
  setCurrentUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  setApplications: (apps: Application[]) => void;
  setCurrentApplication: (app: Application | null) => void;
  setRooms: (rooms: Room[]) => void;
  setCurrentRoom: (room: Room | null) => void;
  addMessage: (roomId: string, message: Message) => void;
  setMessages: (roomId: string, messages: Message[]) => void;
  updateMessage: (roomId: string, messageId: string, updates: Partial<Message>) => void;
  deleteMessage: (roomId: string, messageId: string) => void;
  setTypingUsers: (roomId: string, users: TypingIndicator[]) => void;
  addTypingUser: (roomId: string, indicator: TypingIndicator) => void;
  removeTypingUser: (roomId: string, oderId: string) => void;
  setUserOnline: (userId: string, isOnline: boolean) => void;
  updateRoomLastMessage: (roomId: string, message: Message) => void;
  incrementUnreadCount: (roomId: string) => void;
  clearUnreadCount: (roomId: string) => void;
  logout: () => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  currentUser: null,
  token: localStorage.getItem('chat_token'),
  isAuthenticated: !!localStorage.getItem('chat_token'),
  applications: [],
  currentApplication: null,
  rooms: [],
  currentRoom: null,
  messages: {},
  typingUsers: {},
  onlineUsers: new Set(),

  setCurrentUser: (user) => set({ currentUser: user, isAuthenticated: !!user }),
  
  setToken: (token) => {
    if (token) {
      localStorage.setItem('chat_token', token);
    } else {
      localStorage.removeItem('chat_token');
    }
    set({ token, isAuthenticated: !!token });
  },
  
  setApplications: (applications) => set({ applications }),
  
  setCurrentApplication: (currentApplication) => set({ currentApplication }),
  
  setRooms: (rooms) => set({ rooms }),
  
  setCurrentRoom: (currentRoom) => set({ currentRoom }),
  
  addMessage: (roomId, message) => set((state) => ({
    messages: {
      ...state.messages,
      [roomId]: [...(state.messages[roomId] || []), message],
    },
  })),
  
  setMessages: (roomId, messages) => set((state) => ({
    messages: {
      ...state.messages,
      [roomId]: messages,
    },
  })),
  
  updateMessage: (roomId, messageId, updates) => set((state) => ({
    messages: {
      ...state.messages,
      [roomId]: (state.messages[roomId] || []).map((msg) =>
        msg.id === messageId ? { ...msg, ...updates } : msg
      ),
    },
  })),
  
  deleteMessage: (roomId, messageId) => set((state) => ({
    messages: {
      ...state.messages,
      [roomId]: (state.messages[roomId] || []).map((msg) =>
        msg.id === messageId ? { ...msg, deleted_at: new Date().toISOString() } : msg
      ),
    },
  })),
  
  setTypingUsers: (roomId, users) => set((state) => ({
    typingUsers: {
      ...state.typingUsers,
      [roomId]: users,
    },
  })),
  
  addTypingUser: (roomId, indicator) => set((state) => {
    const existing = state.typingUsers[roomId] || [];
    const filtered = existing.filter((t) => t.user_id !== indicator.user_id);
    return {
      typingUsers: {
        ...state.typingUsers,
        [roomId]: [...filtered, indicator],
      },
    };
  }),
  
  removeTypingUser: (roomId, oderId) => set((state) => ({
    typingUsers: {
      ...state.typingUsers,
      [roomId]: (state.typingUsers[roomId] || []).filter((t) => t.user_id !== oderId),
    },
  })),
  
  setUserOnline: (userId, isOnline) => set((state) => {
    const newOnlineUsers = new Set(state.onlineUsers);
    if (isOnline) {
      newOnlineUsers.add(userId);
    } else {
      newOnlineUsers.delete(userId);
    }
    return { onlineUsers: newOnlineUsers };
  }),
  
  updateRoomLastMessage: (roomId, message) => set((state) => ({
    rooms: state.rooms.map((room) =>
      room.id === roomId ? { ...room, last_message: message } : room
    ),
  })),
  
  incrementUnreadCount: (roomId) => set((state) => ({
    rooms: state.rooms.map((room) =>
      room.id === roomId ? { ...room, unread_count: (room.unread_count || 0) + 1 } : room
    ),
  })),
  
  clearUnreadCount: (roomId) => set((state) => ({
    rooms: state.rooms.map((room) =>
      room.id === roomId ? { ...room, unread_count: 0 } : room
    ),
  })),
  
  logout: () => {
    localStorage.removeItem('chat_token');
    set({
      currentUser: null,
      token: null,
      isAuthenticated: false,
      applications: [],
      currentApplication: null,
      rooms: [],
      currentRoom: null,
      messages: {},
      typingUsers: {},
      onlineUsers: new Set(),
    });
  },
}));

export interface User {
  id: string;
  application_id: string;
  external_id?: string;
  username: string;
  email?: string;
  avatar_url?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  is_online?: boolean;
}

export interface Room {
  id: string;
  application_id: string;
  name: string;
  type: 'group' | 'direct' | 'channel';
  description?: string;
  created_by?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  last_message?: Message;
  unread_count?: number;
  members?: RoomMember[];
}

export interface RoomMember {
  id: string;
  room_id: string;
  user_id: string;
  role: string;
  user?: User;
  joined_at: string;
}

export interface Message {
  id: string;
  room_id: string;
  user_id: string;
  content: string;
  message_type: 'text' | 'image' | 'file' | 'system';
  reply_to?: string;
  reply_to_message?: Message;
  edited_at?: string;
  deleted_at?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  user?: User;
  reactions?: MessageReaction[];
  read_by?: ReadReceipt[];
  status?: 'sending' | 'sent' | 'delivered' | 'read';
}

export interface MessageReaction {
  id: string;
  message_id: string;
  user_id: string;
  reaction: string;
  created_at: string;
  user?: User;
}

export interface TypingIndicator {
  id: string;
  room_id: string;
  user_id: string;
  expires_at: string;
  created_at: string;
  user?: User;
}

export interface ReadReceipt {
  id: string;
  message_id: string;
  user_id: string;
  read_at: string;
  user?: User;
}

export interface Application {
  id: string;
  name: string;
  api_key: string;
  secret_key: string;
  created_at: string;
  updated_at: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
}

export interface WebSocketMessage {
  type: 'message' | 'typing' | 'subscribe' | 'unsubscribe' | 'ping' | 'pong' | 'reaction' | 'read';
  room_id?: string;
  payload?: unknown;
}

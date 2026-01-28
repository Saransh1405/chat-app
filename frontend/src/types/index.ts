export interface User {
  id: string;
  application_id?: string;
  external_id?: string;
  username: string;
  email?: string;
  avatar_url?: string;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface Application {
  id: string;
  name: string;
  api_key: string;
  secret_key?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface Room {
  id: string;
  application_id?: string;
  name: string;
  type: 'group' | 'direct' | 'channel';
  description?: string;
  created_by?: string;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface RoomMember {
  id: string;
  room_id: string;
  user_id: string;
  role: string;
  joined_at: string;
  last_read_at?: string;
}

export interface Message {
  id: string;
  room_id: string;
  user_id: string;
  content: string;
  message_type: 'text' | 'image' | 'file' | 'system';
  reply_to?: string;
  edited_at?: string;
  deleted_at?: string;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface MessageReaction {
  id: string;
  message_id: string;
  user_id: string;
  reaction: string;
  created_at: string;
}

export interface TypingIndicator {
  id: string;
  room_id: string;
  user_id: string;
  expires_at: string;
  created_at: string;
}

export interface AuthResponse {
  message: string;
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Array<{
      field: string;
      issue: string;
      value?: string;
    }>;
    traceId?: string;
    timestamp?: string;
  };
}

export interface WebSocketMessage {
  type: string;
  payload?: any;
  room_id?: string;
}


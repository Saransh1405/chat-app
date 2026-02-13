export interface User {
  id: string;
  username: string;
  email: string;
  avatar_url: string | null;
  application_id: string;
  is_online?: boolean;
}

export interface Room {
  id: string;
  name: string;
  type: "public" | "private" | "direct" | "channel" | "group";
  member_count: number;
  last_message?: string;
  last_message_at?: string;
  unread_count?: number;
}

export interface Reaction {
  emoji: string;
  count: number;
  users: string[]; // user IDs who reacted
}

export interface Message {
  id: string;
  content: string;
  user_id: string;
  room_id: string;
  created_at?: string;
  reactions?: Reaction[];
  image_url?: string;
  image_name?: string;
  user?: User;
  // File message support
  message_type?: "text" | "image" | "file";
  file?: {
    filename: string;
    file_path: string;
    file_size: number;
    mime_type: string;
  };
}

export interface TypingUser {
  user_id: string;
  username: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface WSEvent {
  type: "message" | "typing" | "presence" | "presence_update" | "reaction";
  payload: unknown;
}

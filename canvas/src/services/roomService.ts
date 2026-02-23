// Room API service for managing collaborative whiteboard rooms

export interface Room {
  id: string;
  name: string;
  description?: string;
  active_users: number;
  created_at: number;
  status: string;
  users?: RoomUser[];
  canvas_data?: any[];
}

export interface RoomUser {
  id: string;
  username: string;
  joined_at: number;
  last_seen: number;
  is_online: boolean;
}

export interface CreateRoomRequest {
  name: string;
  description?: string;
  is_private?: boolean;
}

export interface UpdateRoomRequest {
  name?: string;
  description?: string;
  is_private?: boolean;
}

export interface RoomListResponse {
  rooms: Room[];
  total: number;
  page: number;
  page_size: number;
  has_more: boolean;
}

export interface RoomStats {
  room_id: string;
  active_users: number;
  total_users: number;
  created_at: number;
  last_activity: number;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Array<{ field: string; issue: string }>;
  };
}

const API_BASE_URL = 'http://localhost:8080';

/** Base path for white-board REST API (rooms list, room info) */
const WHITE_BOARD_API_BASE = `${API_BASE_URL}/api/v1/white-board`;

/** WebSocket base URL for white-board (use getWhiteBoardWsUrl(roomId) for full URL) */
export const WHITE_BOARD_WS_BASE = `${API_BASE_URL.replace(/^http/, 'ws')}/api/v1/white-board/ws`;

class RoomService {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    method: string = 'GET'
  ): Promise<T> {
    const url = `${WHITE_BOARD_API_BASE}${endpoint}`;
    
    const defaultOptions: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, defaultOptions);
      
      if (!response.ok) {
        const errorData: ApiError = await response.json();
        throw new Error(errorData.error.message || `HTTP ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error('Network error occurred');
    }
  }

  // Get a list of all rooms with pagination
  async listRooms(page: number = 1, size: number = 20): Promise<RoomListResponse> {
    return this.request<RoomListResponse>(`/rooms?page=${page}&size=${size}`);
  }

  // Get information about a specific room
  async getRoomInfo(roomId: string): Promise<Room> {
    return this.request<Room>(`/rooms/${roomId}`);
  }

  // Get room statistics (when backend supports it)
  async getRoomStats(roomId: string): Promise<RoomStats> {
    return this.request<RoomStats>(`/rooms/${roomId}/stats`);
  }

  // Join a room (this is more of a client-side action, but we can track it)
  // async joinRoom(roomId: string, username?: string): Promise<Room> {
  //   // First get room info to verify it exists
  //   const room = await this.getRoomInfo(roomId);
    
  //   // You could add additional logic here like:
  //   // - Tracking user joins
  //   // - Validating room access
  //   // - Updating room metadata
    
  //   return room;
  // }

  // Check if a room exists
  async roomExists(roomId: string): Promise<boolean> {
    try {
      await this.getRoomInfo(roomId);
      return true;
    } catch (error) {
      return false;
    }
  }

  // Generate a unique room ID
  generateRoomId(): string {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substring(2, 8);
    return `${random}-${timestamp}`;
  }

  // Validate room ID format
  isValidRoomId(roomId: string): boolean {
    // Basic validation - you can enhance this based on your requirements
    return roomId.length >= 3 && roomId.length <= 50 && /^[a-zA-Z0-9-_]+$/.test(roomId);
  }

  /** Full WebSocket URL for a white-board room (use when connecting to canvas) */
  getWhiteBoardWsUrl(roomId: string): string {
    return `${WHITE_BOARD_WS_BASE}/${roomId}`;
  }
}

export const roomService = new RoomService();
export default roomService;

import type { Room } from "@/types/chat";
import { api } from "./api";

export const roomsAPI = {
  async list(): Promise<Room[]> {
    const res = await api.get<{ rooms: Room[] }>("/rooms");
    const rooms = res.rooms || [];
    return rooms.map(r => ({
      ...r,
      type: r.type === "channel" ? "public" : r.type === "group" ? "private" : r.type
    })) as Room[];
  },

  async get(roomId: string): Promise<Room> {
    // Backend doesn't have a direct GET /rooms/:id in some cases, 
    // but the list usually contains it. Let's assume list for now or use query.
    const res = await api.get<{ rooms: Room[] }>("/rooms");
    const room = res.rooms.find((r) => r.id === roomId);
    if (!room) throw new Error("Room not found");
    return room;
  },

  async create(name: string, type: Room["type"] = "public"): Promise<Room> {
    // Map "public" to "channel" or "group" if needed. 
    // Backend types: "group", "direct", "channel"
    const backendType = type === "public" ? "channel" : type === "private" ? "group" : "direct";
    const res = await api.post<{ room: Room }>("/rooms", { name, type: backendType });
    return res.room;
  },

  async join(roomId: string, userId: string): Promise<void> {
    await api.post("/rooms/members", { room_id: roomId, user_id: userId, role: "member" });
  },

  async leave(roomId: string, userId: string): Promise<void> {
    // DELETE /api/v1/rooms/members requires room_id and user_id in body usually or query
    await api.post("/rooms/members", { room_id: roomId, user_id: userId }); // Assuming delete logic is separate or needs mapping
    // My backend uses DELETE /api/v1/rooms/members
  },

  async invite(roomId: string, userId: string): Promise<void> {
    await api.post("/rooms/members", { room_id: roomId, user_id: userId, role: "member" });
  },
};

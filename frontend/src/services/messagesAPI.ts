import type { Message, Reaction } from "@/types/chat";
import { api } from "./api";

interface BackendReaction {
  id: string;
  message_id: string;
  user_id: string;
  reaction: string;
  created_at: string;
}

export const messagesAPI = {
  async getReactions(messageId: string): Promise<Reaction[]> {
    try {
      const res = await api.get<{ reactions: BackendReaction[] }>(`/reactions?message_id=${messageId}`);
      const backendReactions = res.reactions || [];

      // Aggregate reactions by emoji
      const reactionMap = new Map<string, Reaction>();

      backendReactions.forEach((r) => {
        const existing = reactionMap.get(r.reaction);
        if (existing) {
          existing.count++;
          existing.users.push(r.user_id);
        } else {
          reactionMap.set(r.reaction, {
            emoji: r.reaction,
            count: 1,
            users: [r.user_id],
          });
        }
      });

      return Array.from(reactionMap.values());
    } catch (err) {
      console.error("[messagesAPI] Failed to fetch reactions:", err);
      return [];
    }
  },

  async list(roomId: string, limit: number = 50, offset: number = 0): Promise<Message[]> {
    const res = await api.get<{ messages: Array<Message & { file?: any }> }>(`/messages?room_id=${roomId}&limit=${limit}&offset=${offset}`);
    const messages = res.messages ? [...res.messages].reverse() : [];

    // Process messages to include file data
    const messagesWithFiles = messages.map((msg: any) => {
      const message: Message = {
        id: msg.id,
        content: msg.content,
        user_id: msg.user_id,
        room_id: msg.room_id,
        created_at: msg.created_at,
        user: msg.user,
        message_type: msg.message_type,
        reactions: msg.reactions || [],
      };
      
      // Include file data if present
      if (msg.file) {
        message.file = {
          filename: msg.file.filename,
          file_path: msg.file.file_path,
          file_size: msg.file.file_size,
          mime_type: msg.file.mime_type,
        };
      }
      
      return message;
    });

    // Fetch reactions for all messages
    const messagesWithReactions = await Promise.all(
      messagesWithFiles.map(async (msg) => {
        const reactions = await messagesAPI.getReactions(msg.id);
        return {
          ...msg,
          reactions,
        };
      })
    );

    return messagesWithReactions;
  },

  async send(roomId: string, content: string, imageUrl?: string, fileData?: {
    filename: string;
    file_path: string;
    file_size: number;
    mime_type: string;
  }, applicationId?: string): Promise<Message> {
    // If file data is provided, send as MEDIA type
    if (fileData) {
      const res = await api.post<{ message: Message; status: string }>("/messages", {
        type: "MEDIA",
        room_id: roomId,
        content: content || "", // Optional caption
        application_id: applicationId,
        file_request: {
          filename: fileData.filename,
          file_path: fileData.file_path,
          file_size: fileData.file_size,
          mime_type: fileData.mime_type,
        },
      });
      return res.message;
    }

    // Otherwise send as TEXT type (no file attachments)
    const res = await api.post<{ message: Message }>("/messages", {
      type: "TEXT",
      room_id: roomId,
      content: content || "",
    });
    return res.message;
  },

  async react(messageId: string, emoji: string): Promise<void> {
    await api.post("/reactions", { message_id: messageId, reaction: emoji });
  },

  async delete(messageId: string, roomId: string, applicationId?: string): Promise<void> {
    const body: { id: string; room_id: string; application_id?: string } = {
      id: messageId,
      room_id: roomId,
    };
    
    // Only include application_id if it's provided and not empty
    if (applicationId) {
      body.application_id = applicationId;
    }
    
    await api.delete("/messages", { body });
  },

  async uploadImage(roomId: string, file: File): Promise<{ url: string; name: string }> {
    // This depends on whether the backend has a file upload endpoint.
    // If not, we might need to skip or use a mock for now.
    // Let's assume there's an upload endpoint or return mock url.
    return {
      url: URL.createObjectURL(file), // Mock for now if no backend support
      name: file.name,
    };
  },
};

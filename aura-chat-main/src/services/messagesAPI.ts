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

  async list(roomId: string): Promise<Message[]> {
    const res = await api.get<{ messages: Message[] }>(`/messages?room_id=${roomId}`);
    const messages = res.messages ? [...res.messages].reverse() : [];

    // Fetch reactions for all messages
    const messagesWithReactions = await Promise.all(
      messages.map(async (msg) => {
        const reactions = await messagesAPI.getReactions(msg.id);
        return {
          ...msg,
          reactions,
        };
      })
    );

    return messagesWithReactions;
  },

  async send(roomId: string, content: string, imageUrl?: string): Promise<Message> {
    const res = await api.post<{ message: Message }>("/messages", {
      room_id: roomId,
      content,
      message_type: imageUrl ? "image" : "text",
      metadata: imageUrl ? { image_url: imageUrl } : {},
    });
    return res.message;
  },

  async react(messageId: string, emoji: string): Promise<void> {
    await api.post("/reactions", { message_id: messageId, reaction: emoji });
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

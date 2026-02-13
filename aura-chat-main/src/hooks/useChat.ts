import { useState, useEffect, useCallback, useRef } from "react";
import type { Room, Message, TypingUser } from "@/types/chat";
import { roomsAPI } from "@/services/roomsAPI";
import { messagesAPI } from "@/services/messagesAPI";
import { wsClient } from "@/services/wsClient";
import { useAuth } from "@/contexts/AuthContext";

export function useChat() {
  const { user } = useAuth();
  const [rooms, setRooms] = useState<Room[]>([]);
  const [activeRoomId, setActiveRoomId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [typingUsers, setTypingUsers] = useState<TypingUser[]>([]);
  const [userPresence, setUserPresence] = useState<Map<string, 'online' | 'offline'>>(new Map());
  const [isLoadingRooms, setIsLoadingRooms] = useState(true);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Load rooms
  useEffect(() => {
    roomsAPI.list().then((data) => {
      setRooms(data);
      if (data.length > 0 && !activeRoomId) {
        setActiveRoomId(data[0].id);
      }
      setIsLoadingRooms(false);
    });
  }, []);

  // Load messages when active room changes
  useEffect(() => {
    if (!activeRoomId) return;
    setIsLoadingMessages(true);
    setTypingUsers([]);
    messagesAPI.list(activeRoomId).then((data) => {
      setMessages(data);
      setIsLoadingMessages(false);
    });
  }, [activeRoomId]);

  // WebSocket listeners
  useEffect(() => {
    if (!activeRoomId) return;

    console.log("[useChat] Setting up WebSocket listeners for room:", activeRoomId);

    // Subscribe to room
    wsClient.send("subscribe", {}, activeRoomId);

    const unsubMessage = wsClient.on("message", (payload: any) => {
      console.log("[useChat] Received message event:", payload);
      // Backend sends: { message: {...}, file: {...} }
      const message = payload.message || payload;
      const file = payload.file;
      
      // Merge file data into message if present
      const messageWithFile = file ? {
        ...message,
        file: {
          filename: file.filename,
          file_path: file.file_path,
          file_size: file.file_size,
          mime_type: file.mime_type,
        }
      } : message;
      
      if (messageWithFile.room_id === activeRoomId) {
        console.log("[useChat] Message belongs to active room, adding to state");
        setMessages((prev) => {
          if (prev.some(m => m.id === messageWithFile.id)) {
            console.log("[useChat] Message already exists, skipping");
            return prev;
          }
          console.log("[useChat] Adding new message to state");
          return [...prev, messageWithFile];
        });
      } else {
        console.log("[useChat] Message belongs to different room:", messageWithFile.room_id);
      }
    });

    const unsubTyping = wsClient.on("typing", (payload: any) => {
      console.log("[useChat] Received typing event:", payload);
      if (payload.room_id === activeRoomId && payload.user_id !== user?.id) {
        setTypingUsers((prev) => {
          if (prev.find((t) => t.user_id === payload.user_id)) return prev;
          return [...prev, { user_id: payload.user_id, username: payload.username }];
        });
        // Remove after 3s
        setTimeout(() => {
          setTypingUsers((prev) => prev.filter((t) => t.user_id !== payload.user_id));
        }, 3000);
      }
    });

    const unsubReactionAdded = wsClient.on("reaction_added", (payload: any) => {
      console.log("[useChat] Received reaction_added event:", payload);
      console.log("[useChat] Current activeRoomId:", activeRoomId);
      console.log("[useChat] Current messages count:", messages.length);

      // payload contains: { id, message_id, user_id, reaction, created_at }
      setMessages((prev) => {
        console.log("[useChat] Updating messages for reaction_added, prev count:", prev.length);
        const updated = prev.map((msg) => {
          if (msg.id !== payload.message_id) return msg;

          console.log("[useChat] Found message to update:", msg.id);
          const reactions = msg.reactions || [];
          console.log("[useChat] Current reactions:", reactions);
          const existingReaction = reactions.find((r) => r.emoji === payload.reaction);

          if (existingReaction) {
            // Add user to existing reaction
            if (!existingReaction.users.includes(payload.user_id)) {
              console.log("[useChat] Adding user to existing reaction");
              return {
                ...msg,
                reactions: reactions.map((r) =>
                  r.emoji === payload.reaction
                    ? { ...r, count: r.count + 1, users: [...r.users, payload.user_id] }
                    : r
                ),
              };
            } else {
              console.log("[useChat] User already has this reaction");
            }
          } else {
            // Create new reaction
            console.log("[useChat] Creating new reaction");
            return {
              ...msg,
              reactions: [
                ...reactions,
                { emoji: payload.reaction, count: 1, users: [payload.user_id] },
              ],
            };
          }
          return msg;
        });
        console.log("[useChat] Updated messages count:", updated.length);
        return updated;
      });
    });

    const unsubReactionRemoved = wsClient.on("reaction_removed", (payload: any) => {
      console.log("[useChat] Received reaction_removed event:", payload);
      // payload contains: { message_id, user_id, reaction }
      setMessages((prev) =>
        prev.map((msg) => {
          if (msg.id !== payload.message_id) return msg;

          const reactions = msg.reactions || [];
          return {
            ...msg,
            reactions: reactions
              .map((r) =>
                r.emoji === payload.reaction
                  ? { ...r, count: r.count - 1, users: r.users.filter((u) => u !== payload.user_id) }
                  : r
              )
              .filter((r) => r.count > 0),
          };
        })
      );
    });

    const unsubPresenceUpdate = wsClient.on("presence_update", (payload: any) => {
      console.log("[useChat] Received presence_update event:", payload);
      // payload contains: { user_id, status, timestamp }
      // Don't update presence for current user (we know we're online)
      if (payload.user_id === user?.id) {
        return;
      }
      
      setUserPresence((prev) => {
        const updated = new Map(prev);
        updated.set(payload.user_id, payload.status);
        return updated;
      });
    });

    return () => {
      console.log("[useChat] Cleaning up WebSocket listeners for room:", activeRoomId);
      wsClient.send("unsubscribe", {}, activeRoomId);
      unsubMessage();
      unsubTyping();
      unsubReactionAdded();
      unsubReactionRemoved();
      unsubPresenceUpdate();
    };
  }, [activeRoomId, user?.id]);

  const sendMessage = useCallback(
    async (content: string, imageUrl?: string, fileData?: {
      filename: string;
      file_path: string;
      file_size: number;
      mime_type: string;
    }) => {
      if (!activeRoomId) return;
      
      // Backend broadcasts the message, so we don't need to send it via WS manually.
      // We might want to add it to local state immediately for better UX (optimistic update),
      // but the WS event will also trigger an update.
      
      // messagesAPI.send() will automatically set:
      // - type: "MEDIA" if fileData is provided
      // - type: "TEXT" if no fileData
      await messagesAPI.send(activeRoomId, content, imageUrl, fileData, user?.application_id);
      
      // Let the WebSocket event update the message list to ensure consistency
    },
    [activeRoomId, user?.application_id]
  );

  const sendTyping = useCallback(() => {
    if (!activeRoomId || !user) return;

    // Throttle typing events to max once per second
    if (typingTimeoutRef.current) return;

    console.log("[useChat] Sending typing indicator");

    // Send typing indicator via REST API
    fetch(`${import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1"}/typing`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${localStorage.getItem("access_token")}`,
      },
      body: JSON.stringify({
        room_id: activeRoomId,
        user_id: user.id,
      }),
    }).catch(err => console.error("[useChat] Failed to send typing indicator:", err));

    // Throttle for 1 second
    typingTimeoutRef.current = setTimeout(() => {
      typingTimeoutRef.current = null;
    }, 1000);
  }, [activeRoomId, user]);

  const addReaction = useCallback(
    async (messageId: string, emoji: string) => {
      console.log("[useChat] Adding reaction:", { messageId, emoji });
      // Just call the API - the WebSocket event will update the UI for everyone including the sender
      await messagesAPI.react(messageId, emoji);
    },
    []
  );

  const createRoom = useCallback(async (name: string, type: Room["type"]) => {
    const room = await roomsAPI.create(name, type);
    setRooms((prev) => [...prev, room]);
    setActiveRoomId(room.id);
  }, []);

  const activeRoom = rooms.find((r) => r.id === activeRoomId) || null;

  return {
    rooms,
    activeRoom,
    activeRoomId,
    setActiveRoomId,
    messages,
    typingUsers,
    userPresence,
    isLoadingRooms,
    isLoadingMessages,
    sendMessage,
    sendTyping,
    addReaction,
    createRoom,
  };
}

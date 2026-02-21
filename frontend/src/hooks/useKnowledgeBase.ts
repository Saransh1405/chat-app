import { useState, useCallback, useRef } from "react";
import { type Message, type User } from "@/types/chat";
import { api } from "@/services/api";
import { useAuth } from "@/contexts/AuthContext";

interface ChatMessage {
    id: string;
    session_id: string;
    role: "user" | "assistant";
    content: string;
    timestamp: number;
}

export function useKnowledgeBase() {
    const { user } = useAuth();
    const [messages, setMessages] = useState<Message[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [sessionId, setSessionId] = useState<string | null>(null);
    const [isTyping, setIsTyping] = useState(false);
    const eventSourceRef = useRef<EventSource | null>(null);

    const convertToFrontendMessage = (msg: ChatMessage): Message => {
        const isUser = msg.role === "user";
        return {
            id: msg.id,
            content: msg.content,
            user_id: isUser ? user?.id || "user" : "ai-assistant",
            room_id: "knowledge-base",
            created_at: new Date(msg.timestamp * 1000).toISOString(),
            message_type: "text",
            user: isUser ? user! : {
                id: "ai-assistant",
                username: "AI Assistant",
                email: "ai@aura.chat",
                avatar_url: "",
                application_id: "",
                created_at: "",
                updated_at: ""
            } as User,
            reactions: []
        };
    };

    const loadHistory = useCallback(async (sid: string) => {
        try {
            setIsLoading(true);
            const res = await api.get<{ messages: ChatMessage[] }>(`/answer/history?sessionID=${sid}`);
            if (res.messages) {
                setMessages(res.messages.map(convertToFrontendMessage));
            }
        } catch (error) {
            console.error("Failed to load history:", error);
        } finally {
            setIsLoading(false);
        }
    }, [user]);

    const sendMessage = useCallback(async (content: string) => {
        if (!content.trim()) return;

        // Optimistic user message
        const tempUserMsg: Message = {
            id: crypto.randomUUID(),
            content: content,
            user_id: user?.id || "user",
            room_id: "knowledge-base",
            created_at: new Date().toISOString(),
            message_type: "text",
            user: user!,
            reactions: []
        };
        setMessages(prev => [...prev, tempUserMsg]);

        // Optimistic AI message (placeholder)
        const tempAiMsgId = crypto.randomUUID();
        const tempAiMsg: Message = {
            id: tempAiMsgId,
            content: "",
            user_id: "ai-assistant",
            room_id: "knowledge-base",
            created_at: new Date().toISOString(),
            message_type: "text",
            user: {
                id: "ai-assistant",
                username: "AI Assistant",
                email: "ai@aura.chat",
                avatar_url: "",
                application_id: "",
                created_at: "",
                updated_at: ""
            } as User,
            reactions: []
        };
        setMessages(prev => [...prev, tempAiMsg]);

        try {
            console.log("Calling answer API with question:", content, "sessionId:", sessionId);
            const res = await api.post<{ sessionId: string, message: string }>("/answer", {
                question: content,
                sessionId: sessionId ? sessionId : undefined
            });

            console.log("Answer API response:", res);
            const newSessionId = res.sessionId;
            if (!sessionId) {
                setSessionId(newSessionId);
            }

            // Start SSE stream for real-time response
            if (eventSourceRef.current) {
                eventSourceRef.current.close();
            }

            const baseUrl = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";
            const token = localStorage.getItem("access_token");
            const streamUrl = `${baseUrl}/answer/stream?sessionID=${newSessionId}&token=${encodeURIComponent(token || "")}`;
            console.log("Connecting to SSE stream:", streamUrl);
            const evtSource = new EventSource(streamUrl);
            eventSourceRef.current = evtSource;

            evtSource.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    
                    if (data.type === "typing") {
                        // Show typing indicator
                        setIsTyping(true);
                    } else if (data.type === "content") {
                        // Hide typing indicator and show content
                        setIsTyping(false);
                        const chunk = data.content;

                        setMessages(prev => prev.map(msg => {
                            if (msg.id === tempAiMsgId) {
                                return { ...msg, content: msg.content + chunk };
                            }
                            return msg;
                        }));
                    } else {
                        // Backward compatibility: if no type, assume it's content
                        const chunk = data.content;
                        setIsTyping(false);
                        setMessages(prev => prev.map(msg => {
                            if (msg.id === tempAiMsgId) {
                                return { ...msg, content: msg.content + chunk };
                            }
                            return msg;
                        }));
                    }
                } catch (parseError) {
                    console.error("Error parsing SSE message:", parseError, event.data);
                }
            };

            evtSource.onerror = (error) => {
                console.error("SSE stream error:", error);
                evtSource.close();
                eventSourceRef.current = null;
            };

        } catch (error) {
            console.error("Failed to send message to answer API:", error);
            // Remove only the AI placeholder message, keep the user message
            setMessages(prev => prev.map(msg => {
                if (msg.id === tempAiMsgId) {
                    // Replace with error message
                    return {
                        ...msg,
                        content: "Sorry, I encountered an error while processing your question. Please try again."
                    };
                }
                return msg;
            }));
        }
    }, [sessionId, user, loadHistory]);

    return {
        messages,
        isLoading,
        isTyping,
        sendMessage,
        loadHistory,
        sessionId
    };
}

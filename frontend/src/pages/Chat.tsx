import { useState } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useChat } from "@/hooks/useChat";
import { useKnowledgeBase } from "@/hooks/useKnowledgeBase";
import { Sidebar } from "@/components/chat/Sidebar";
import { RoomHeader } from "@/components/chat/RoomHeader";
import { MessageList } from "@/components/chat/MessageList";
import { MessageInput } from "@/components/chat/MessageInput";
import { TypingIndicator } from "@/components/chat/TypingIndicator";

const Chat = () => {
    const { user, logout } = useAuth();
    const [mode, setMode] = useState<"chat" | "knowledge_base">("chat");
    const [sidebarOpen, setSidebarOpen] = useState(false);

    const {
        rooms,
        activeRoom,
        activeRoomId,
        setActiveRoomId,
        messages: chatMessages,
        typingUsers,
        userPresence,
        isLoadingMessages: isChatLoading,
        sendMessage: sendChatMessage,
        sendTyping,
        addReaction,
        deleteMessage,
        createRoom,
    } = useChat();

    const {
        messages: kbMessages,
        isLoading: isKbLoading,
        isTyping: isKbTyping,
        sendMessage: sendKbMessage,
    } = useKnowledgeBase();

    const handleModeChange = (newMode: "chat" | "knowledge_base") => {
        setMode(newMode);
    };

    const currentMessages = mode === "chat" ? chatMessages : kbMessages;
    const isLoading = mode === "chat" ? isChatLoading : isKbLoading;

    return (
        <div className="flex h-screen w-full overflow-hidden bg-background">
            <Sidebar
                rooms={rooms}
                activeRoomId={activeRoomId}
                currentUser={user}
                onSelectRoom={setActiveRoomId}
                onCreateRoom={createRoom}
                onLogout={logout}
                isOpen={sidebarOpen}
                onToggle={() => setSidebarOpen((prev) => !prev)}
                mode={mode}
                onModeChange={handleModeChange}
            />

            <main className="flex flex-1 flex-col min-w-0">
                <RoomHeader room={activeRoom} mode={mode} />

                <MessageList
                    messages={currentMessages}
                    currentUserId={user?.id || "user-1"}
                    isLoading={isLoading}
                    onReact={mode === "chat" ? addReaction : () => { }}
                    onDelete={mode === "chat" ? deleteMessage : undefined}
                    userPresence={mode === "chat" ? userPresence : undefined}
                />

                {mode === "chat" && <TypingIndicator typingUsers={typingUsers} />}
                {mode === "knowledge_base" && isKbTyping && (
                    <div className="px-4 py-2">
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                            <div className="flex gap-0.5">
                                {[0, 1, 2].map((i) => (
                                    <span
                                        key={i}
                                        className="inline-block h-1.5 w-1.5 rounded-full bg-primary animate-typing-dot"
                                        style={{ animationDelay: `${i * 0.2}s` }}
                                    />
                                ))}
                            </div>
                            <span>AI Assistant is thinking...</span>
                        </div>
                    </div>
                )}

                <MessageInput
                    onSend={mode === "chat" ? sendChatMessage : (content) => sendKbMessage(content)}
                    onTyping={mode === "chat" ? sendTyping : () => { }}
                    disabled={mode === "chat" ? !activeRoom : false}
                    mode={mode}
                />
            </main>
        </div>
    );
};

export default Chat;

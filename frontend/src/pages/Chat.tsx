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
        createRoom,
    } = useChat();

    const {
        messages: kbMessages,
        isLoading: isKbLoading,
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
                    userPresence={mode === "chat" ? userPresence : undefined}
                />

                {mode === "chat" && <TypingIndicator typingUsers={typingUsers} />}

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

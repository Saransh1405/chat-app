import { useState } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useChat } from "@/hooks/useChat";
import { Sidebar } from "@/components/chat/Sidebar";
import { RoomHeader } from "@/components/chat/RoomHeader";
import { MessageList } from "@/components/chat/MessageList";
import { MessageInput } from "@/components/chat/MessageInput";
import { TypingIndicator } from "@/components/chat/TypingIndicator";

const Chat = () => {
  const { user, logout } = useAuth();
  const {
    rooms,
    activeRoom,
    activeRoomId,
    setActiveRoomId,
    messages,
    typingUsers,
    isLoadingMessages,
    sendMessage,
    sendTyping,
    addReaction,
    createRoom,
  } = useChat();

  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="flex h-screen w-full overflow-hidden bg-background">
      {/* Sidebar */}
      <Sidebar
        rooms={rooms}
        activeRoomId={activeRoomId}
        currentUser={user}
        onSelectRoom={setActiveRoomId}
        onCreateRoom={createRoom}
        onLogout={logout}
        isOpen={sidebarOpen}
        onToggle={() => setSidebarOpen((prev) => !prev)}
      />

      {/* Main chat area */}
      <main className="flex flex-1 flex-col min-w-0">
        <RoomHeader room={activeRoom} />

        <MessageList
          messages={messages}
          currentUserId={user?.id || "user-1"}
          isLoading={isLoadingMessages}
          onReact={addReaction}
        />

        <TypingIndicator typingUsers={typingUsers} />

        <MessageInput
          onSend={sendMessage}
          onTyping={sendTyping}
          disabled={!activeRoom}
        />
      </main>
    </div>
  );
};

export default Chat;

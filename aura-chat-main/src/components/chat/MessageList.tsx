import { useEffect, useRef } from "react";
import type { Message } from "@/types/chat";
import { MessageBubble } from "./MessageBubble";
import { Loader2 } from "lucide-react";

interface MessageListProps {
  messages: Message[];
  currentUserId: string;
  isLoading: boolean;
  onReact: (messageId: string, emoji: string) => void;
}

export function MessageList({ messages, currentUserId, isLoading, onReact }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  if (isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (messages.length === 0) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-2 text-muted-foreground">
        <span className="text-4xl">💬</span>
        <p className="text-sm">No messages yet. Start the conversation!</p>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto scrollbar-thin py-4 space-y-1">
      {messages.map((msg, i) => {
        const isMe = msg.user_id === currentUserId;
        const prevMsg = messages[i - 1];
        const showAvatar = !prevMsg || prevMsg.user_id !== msg.user_id;

        return (
          <MessageBubble
            key={msg.id}
            message={msg}
            isMe={isMe}
            showAvatar={showAvatar}
            currentUserId={currentUserId}
            onReact={onReact}
          />
        );
      })}
      <div ref={bottomRef} />
    </div>
  );
}

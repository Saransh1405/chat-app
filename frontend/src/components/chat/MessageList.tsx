import { useEffect, useRef } from "react";
import type { Message } from "@/types/chat";
import { MessageBubble } from "./MessageBubble";
import { Loader2 } from "lucide-react";

interface MessageListProps {
  messages: Message[];
  currentUserId: string;
  isLoading: boolean;
  onReact: (messageId: string, emoji: string) => void;
  onDelete?: (messageId: string) => void;
  userPresence?: Map<string, 'online' | 'offline'>;
}

export function MessageList({ messages, currentUserId, isLoading, onReact, onDelete, userPresence }: MessageListProps) {
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
      <div className="flex flex-1 flex-col items-center justify-center gap-6 px-6 py-12">
        <div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-primary/10 text-primary">
          <svg className="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" />
          </svg>
        </div>
        <div className="text-center space-y-1">
          <p className="text-base font-medium text-foreground">No messages yet</p>
          <p className="text-sm text-muted-foreground max-w-xs">Send a message to start the conversation.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto scrollbar-thin py-6 px-2 sm:px-4 space-y-2">
      {messages.map((msg, i) => {
        const isMe = msg.user_id === currentUserId;
        const prevMsg = messages[i - 1];
        const showAvatar = !prevMsg || prevMsg.user_id !== msg.user_id;

        const isOnline = msg.user_id === currentUserId ? true : userPresence?.get(msg.user_id) === 'online';

        return (
          <MessageBubble
            key={msg.id}
            message={msg}
            isMe={isMe}
            showAvatar={showAvatar}
            currentUserId={currentUserId}
            onReact={onReact}
            onDelete={onDelete}
            isOnline={isOnline}
          />
        );
      })}
      <div ref={bottomRef} />
    </div>
  );
}

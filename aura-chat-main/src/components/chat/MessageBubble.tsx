import { format } from "date-fns";
import { cn } from "@/lib/utils";
import { UserAvatar } from "./UserAvatar";
import { EmojiPicker } from "./EmojiPicker";
import type { Message, User } from "@/types/chat";
import { motion } from "framer-motion";
import { SmilePlus } from "lucide-react";

interface MessageBubbleProps {
  message: Message;
  isMe: boolean;
  showAvatar: boolean;
  currentUserId: string;
  onReact: (messageId: string, emoji: string) => void;
}

export function MessageBubble({ message, isMe, showAvatar, currentUserId, onReact }: MessageBubbleProps) {
  const sender = message.user;
  const time = message.created_at ? format(new Date(message.created_at), "h:mm a") : "";

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className={cn("group flex gap-3 px-4 py-1", isMe ? "flex-row-reverse" : "flex-row")}
    >
      {/* Avatar */}
      <div className="w-8 flex-shrink-0 pt-1">
        {showAvatar && sender && (
          <UserAvatar
            userId={sender.id}
            username={sender.username}
            avatarUrl={sender.avatar_url}
            size="sm"
          />
        )}
      </div>

      {/* Bubble */}
      <div className={cn("flex max-w-[70%] flex-col", isMe ? "items-end" : "items-start")}>
        {/* Username + time */}
        {showAvatar && (
          <div className={cn("flex items-center gap-2 px-1 pb-1 text-xs text-muted-foreground", isMe && "flex-row-reverse")}>
            <span className="font-medium text-foreground">{isMe ? "You" : sender?.username}</span>
            <span>{time}</span>
          </div>
        )}

        {/* Content */}
        <div
          className={cn(
            "relative rounded-2xl px-4 py-2.5 text-sm leading-relaxed shadow-sm",
            isMe
              ? "gradient-primary text-white rounded-tr-md"
              : "bg-bubble-other text-bubble-other-foreground rounded-tl-md"
          )}
        >
          {message.content && <p>{message.content}</p>}

          {/* Image attachment */}
          {message.image_url && (
            <div className="mt-2 overflow-hidden rounded-lg">
              <img
                src={message.image_url}
                alt={message.image_name || "Attachment"}
                className="max-h-64 w-full object-cover"
                loading="lazy"
              />
              {message.image_name && (
                <p className={cn("mt-1 text-xs", isMe ? "text-white/70" : "text-muted-foreground")}>
                  {message.image_name}
                </p>
              )}
            </div>
          )}

          {/* Add reaction button (appears on hover) */}
          <div
            className={cn(
              "absolute top-1/2 -translate-y-1/2 opacity-0 transition-opacity group-hover:opacity-100",
              isMe ? "-left-8" : "-right-8"
            )}
          >
            <EmojiPicker
              onSelect={(emoji) => onReact(message.id, emoji)}
              trigger={
                <button className="flex h-6 w-6 items-center justify-center rounded-full bg-muted text-muted-foreground shadow-sm hover:bg-secondary">
                  <SmilePlus className="h-3.5 w-3.5" />
                </button>
              }
            />
          </div>
        </div>

        {/* Reactions */}
        {(message.reactions?.length ?? 0) > 0 && (
          <div className={cn("mt-1 flex flex-wrap gap-1 px-1", isMe && "justify-end")}>
            {message.reactions?.map((reaction) => (
              <button
                key={reaction.emoji}
                onClick={() => onReact(message.id, reaction.emoji)}
                className={cn(
                  "flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs transition-colors",
                  reaction.users.includes(currentUserId)
                    ? "border-primary/40 bg-primary/10 text-primary"
                    : "border-border bg-muted/50 text-muted-foreground hover:border-primary/30"
                )}
              >
                <span>{reaction.emoji}</span>
                <span>{reaction.count}</span>
              </button>
            ))}
          </div>
        )}
      </div>
    </motion.div>
  );
}

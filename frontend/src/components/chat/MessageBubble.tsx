import { format } from "date-fns";
import { cn } from "@/lib/utils";
import { UserAvatar } from "./UserAvatar";
import { EmojiPicker } from "./EmojiPicker";
import type { Message, User } from "@/types/chat";
import { motion } from "framer-motion";
import { SmilePlus, File, Trash2 } from "lucide-react";

interface MessageBubbleProps {
    message: Message;
    isMe: boolean;
    showAvatar: boolean;
    currentUserId: string;
    onReact: (messageId: string, emoji: string) => void;
    onDelete?: (messageId: string) => void;
    isOnline?: boolean;
}

export function MessageBubble({ message, isMe, showAvatar, currentUserId, onReact, onDelete, isOnline }: MessageBubbleProps) {
    const sender = message.user;
    const time = message.created_at ? format(new Date(message.created_at), "h:mm a") : "";

    const citationRegex = /\[Source: (.*?)\]/g;
    const citations: string[] = [];
    let displayContent = message.content || "";

    let match;
    while ((match = citationRegex.exec(displayContent)) !== null) {
        if (!citations.includes(match[1])) {
            citations.push(match[1]);
        }
    }

    displayContent = displayContent.replace(citationRegex, "").trim();

    return (
        <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2 }}
            className={cn("group flex gap-3 px-4 py-1", isMe ? "flex-row-reverse" : "flex-row")}
        >
            <div className="w-8 flex-shrink-0 pt-1">
                {showAvatar && sender && (
                    <UserAvatar
                        userId={sender.id}
                        username={sender.username}
                        avatarUrl={sender.avatar_url}
                        size="sm"
                        isOnline={!isMe ? isOnline : undefined}
                    />
                )}
            </div>

            <div className={cn("flex max-w-[70%] flex-col", isMe ? "items-end" : "items-start")}>
                {showAvatar && (
                    <div className={cn("flex items-center gap-2 px-1 pb-1 text-xs text-muted-foreground", isMe && "flex-row-reverse")}>
                        <span className="font-medium text-foreground">{isMe ? "You" : sender?.username}</span>
                        <span>{time}</span>
                        {citations.length > 0 && <File className="h-3 w-3 text-primary" />}
                    </div>
                )}

                <div
                    className={cn(
                        "relative rounded-2xl px-4 py-2.5 text-sm leading-relaxed",
                        isMe
                            ? "gradient-primary text-white rounded-tr-md shadow-lg shadow-primary/20"
                            : "bg-bubble-other text-bubble-other-foreground rounded-tl-md shadow-md border border-border/50"
                    )}
                >
                    {displayContent ? (
                        <p className="whitespace-pre-wrap">{displayContent}</p>
                    ) : !isMe && message.user_id === "ai-assistant" ? (
                        <div className="flex items-center gap-1 py-1">
                            {[0, 1, 2].map((i) => (
                                <span
                                    key={i}
                                    className="inline-block h-2 w-2 rounded-full bg-muted-foreground/50 animate-typing-dot"
                                    style={{ animationDelay: `${i * 0.2}s` }}
                                />
                            ))}
                        </div>
                    ) : null}

                    {citations.length > 0 && (
                        <div className="mt-3 pt-2 border-t border-border/20">
                            <p className="text-xs font-semibold mb-1 opacity-70">Sources:</p>
                            <div className="flex flex-wrap gap-1">
                                {citations.map((source, i) => (
                                    <div key={i} className="flex items-center gap-1 text-xs bg-black/10 dark:bg-white/10 px-2 py-1 rounded-md">
                                        <File className="h-3 w-3" />
                                        <span className="truncate max-w-[150px]">{source}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

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

                    {message.file && (
                        <div className="mt-2">
                            {message.file.mime_type.startsWith("image/") ? (
                                <div className="overflow-hidden rounded-lg">
                                    <a
                                        href={message.file.file_path}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        <img
                                            src={message.file.file_path}
                                            alt={message.file.filename}
                                            className="max-h-64 w-full cursor-pointer object-cover transition-opacity hover:opacity-90"
                                            loading="lazy"
                                        />
                                    </a>
                                    {message.file.filename && (
                                        <p className={cn("mt-1 text-xs", isMe ? "text-white/70" : "text-muted-foreground")}>
                                            {message.file.filename}
                                        </p>
                                    )}
                                </div>
                            ) : (
                                <a
                                    href={message.file.file_path}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className={cn(
                                        "flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-opacity-80",
                                        isMe
                                            ? "border-white/20 bg-white/10"
                                            : "border-border bg-muted/50"
                                    )}
                                >
                                    <div className={cn(
                                        "flex h-10 w-10 items-center justify-center rounded-md",
                                        isMe ? "bg-white/20" : "bg-primary/10"
                                    )}>
                                        <File className={cn(
                                            "h-5 w-5",
                                            isMe ? "text-white" : "text-primary"
                                        )} />
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <p className={cn(
                                            "truncate text-sm font-medium",
                                            isMe ? "text-white" : "text-foreground"
                                        )}>
                                            {message.file.filename}
                                        </p>
                                        <p className={cn(
                                            "text-xs",
                                            isMe ? "text-white/70" : "text-muted-foreground"
                                        )}>
                                            {(message.file.file_size / 1024).toFixed(1)} KB
                                        </p>
                                    </div>
                                </a>
                            )}
                        </div>
                    )}

                    <div
                        className={cn(
                            "absolute top-1/2 -translate-y-1/2 opacity-0 transition-opacity group-hover:opacity-100 flex gap-1",
                            isMe ? "-left-12" : "-right-12"
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
                        {isMe && onDelete && (
                            <button
                                onClick={() => {
                                    if (confirm("Are you sure you want to delete this message?")) {
                                        onDelete(message.id);
                                    }
                                }}
                                className="flex h-6 w-6 items-center justify-center rounded-full bg-destructive/10 text-destructive shadow-sm hover:bg-destructive/20"
                                title="Delete message"
                            >
                                <Trash2 className="h-3.5 w-3.5" />
                            </button>
                        )}
                    </div>
                </div>

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

import { useRef, useEffect, useState, useCallback } from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { MessageBubble } from './MessageBubble';
import { MessageInput } from './MessageInput';
import { ChatHeader } from './ChatHeader';
import { TypingIndicator } from './TypingIndicator';
import { useChatStore } from '@/store/chatStore';
import type { Message, Room } from '@/types/chat';
import { MessageSquare } from 'lucide-react';

interface ChatAreaProps {
  room: Room | null;
}

export function ChatArea({ room }: ChatAreaProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [replyTo, setReplyTo] = useState<Message | null>(null);
  const { messages, currentUser, typingUsers, addMessage, updateMessage, deleteMessage } = useChatStore();

  const roomMessages = room ? messages[room.id] || [] : [];
  const roomTypingUsers = room ? typingUsers[room.id] || [] : [];

  // Filter out current user from typing users
  const typingUsersFiltered = roomTypingUsers.filter(
    (t) => t.user_id !== currentUser?.id && t.user
  );

  // Scroll to bottom on new messages
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [roomMessages.length]);

  const handleSendMessage = useCallback(
    (content: string, type: 'text' | 'image' | 'file' = 'text', metadata?: Record<string, unknown>) => {
      if (!room || !currentUser) return;

      const newMessage: Message = {
        id: crypto.randomUUID(),
        room_id: room.id,
        user_id: currentUser.id,
        content,
        message_type: type,
        reply_to: replyTo?.id,
        reply_to_message: replyTo || undefined,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        user: currentUser,
        status: 'sending',
        metadata,
      };

      addMessage(room.id, newMessage);
      setReplyTo(null);

      // Simulate message sent status
      setTimeout(() => {
        updateMessage(room.id, newMessage.id, { status: 'sent' });
      }, 500);

      // Simulate message delivered status
      setTimeout(() => {
        updateMessage(room.id, newMessage.id, { status: 'delivered' });
      }, 1500);

      // Simulate message read status
      setTimeout(() => {
        updateMessage(room.id, newMessage.id, { status: 'read' });
      }, 3000);
    },
    [room, currentUser, replyTo, addMessage, updateMessage]
  );

  const handleTyping = useCallback(() => {
    // In real app, send typing indicator via WebSocket
    console.log('User is typing...');
  }, []);

  const handleReply = useCallback((message: Message) => {
    setReplyTo(message);
  }, []);

  const handleEdit = useCallback(
    (message: Message) => {
      if (!room) return;
      // In real app, open edit modal
      const newContent = prompt('Edit message:', message.content);
      if (newContent && newContent !== message.content) {
        updateMessage(room.id, message.id, {
          content: newContent,
          edited_at: new Date().toISOString(),
        });
      }
    },
    [room, updateMessage]
  );

  const handleDelete = useCallback(
    (message: Message) => {
      if (!room) return;
      if (confirm('Delete this message?')) {
        deleteMessage(room.id, message.id);
      }
    },
    [room, deleteMessage]
  );

  const handleAddReaction = useCallback(
    (messageId: string, emoji: string) => {
      if (!room || !currentUser) return;
      updateMessage(room.id, messageId, {
        reactions: [
          ...(roomMessages.find((m) => m.id === messageId)?.reactions || []),
          {
            id: crypto.randomUUID(),
            message_id: messageId,
            user_id: currentUser.id,
            reaction: emoji,
            created_at: new Date().toISOString(),
          },
        ],
      });
    },
    [room, currentUser, roomMessages, updateMessage]
  );

  const handleRemoveReaction = useCallback(
    (reactionId: string) => {
      if (!room) return;
      // Find and update the message with the reaction removed
      const message = roomMessages.find((m) =>
        m.reactions?.some((r) => r.id === reactionId)
      );
      if (message) {
        updateMessage(room.id, message.id, {
          reactions: message.reactions?.filter((r) => r.id !== reactionId),
        });
      }
    },
    [room, roomMessages, updateMessage]
  );

  if (!room) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center bg-background/50">
        <div className="p-6 rounded-full bg-muted/50 mb-6">
          <MessageSquare className="h-16 w-16 text-muted-foreground" />
        </div>
        <h2 className="text-2xl font-semibold mb-2">Welcome to ChatSDK</h2>
        <p className="text-muted-foreground text-center max-w-md">
          Select a conversation from the sidebar or start a new one to begin messaging
        </p>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col bg-background/50">
      <ChatHeader
        room={room}
        membersCount={room.members?.length}
        isTyping={typingUsersFiltered.length > 0}
        typingUserName={typingUsersFiltered[0]?.user?.username}
      />

      <ScrollArea className="flex-1" ref={scrollRef}>
        <div className="p-4 space-y-4">
          {roomMessages.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <p className="text-muted-foreground">No messages yet</p>
              <p className="text-sm text-muted-foreground mt-1">
                Be the first to send a message!
              </p>
            </div>
          ) : (
            roomMessages.map((message, index) => {
              const prevMessage = roomMessages[index - 1];
              const showAvatar =
                !prevMessage ||
                prevMessage.user_id !== message.user_id ||
                new Date(message.created_at).getTime() -
                  new Date(prevMessage.created_at).getTime() >
                  5 * 60 * 1000;

              return (
                <MessageBubble
                  key={message.id}
                  message={message}
                  isOwn={message.user_id === currentUser?.id}
                  showAvatar={showAvatar}
                  currentUserId={currentUser?.id || ''}
                  onReply={handleReply}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                  onAddReaction={handleAddReaction}
                  onRemoveReaction={handleRemoveReaction}
                />
              );
            })
          )}
        </div>
      </ScrollArea>

      {/* Typing indicator */}
      {typingUsersFiltered.length > 0 && (
        <div className="px-4 py-2 border-t border-border/50">
          <TypingIndicator
            users={typingUsersFiltered.map((t) => ({ username: t.user?.username || 'User' }))}
          />
        </div>
      )}

      <MessageInput
        onSend={handleSendMessage}
        onTyping={handleTyping}
        replyTo={replyTo}
        onCancelReply={() => setReplyTo(null)}
      />
    </div>
  );
}

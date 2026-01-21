import { useState } from 'react';
import { cn } from '@/lib/utils';
import { UserAvatar } from './UserAvatar';
import { ReadReceipt } from './ReadReceipt';
import { MessageReactions } from './MessageReactions';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import { MoreHorizontal, Reply, Pencil, Trash2, FileIcon, Download, Image as ImageIcon } from 'lucide-react';
import { format } from 'date-fns';
import type { Message } from '@/types/chat';

interface MessageBubbleProps {
  message: Message;
  isOwn: boolean;
  showAvatar?: boolean;
  currentUserId: string;
  onReply?: (message: Message) => void;
  onEdit?: (message: Message) => void;
  onDelete?: (message: Message) => void;
  onAddReaction?: (messageId: string, emoji: string) => void;
  onRemoveReaction?: (reactionId: string) => void;
}

export function MessageBubble({
  message,
  isOwn,
  showAvatar = true,
  currentUserId,
  onReply,
  onEdit,
  onDelete,
  onAddReaction,
  onRemoveReaction,
}: MessageBubbleProps) {
  const [showActions, setShowActions] = useState(false);

  const isDeleted = !!message.deleted_at;
  const isEdited = !!message.edited_at;

  const renderContent = () => {
    if (isDeleted) {
      return <span className="italic text-muted-foreground">This message was deleted</span>;
    }

    switch (message.message_type) {
      case 'image':
        return (
          <div className="space-y-2">
            <div className="relative rounded-lg overflow-hidden max-w-xs">
              <img src={message.content} alt="Shared image" className="w-full h-auto" />
              <div className="absolute inset-0 bg-black/0 hover:bg-black/10 transition-colors" />
            </div>
          </div>
        );
      case 'file':
        const fileName = message.metadata?.file_name as string || 'File';
        const fileSize = message.metadata?.file_size as number;
        return (
          <div className="flex items-center gap-3 p-3 rounded-lg bg-background/50">
            <div className="p-2 rounded-lg bg-primary/10">
              <FileIcon className="h-6 w-6 text-primary" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="font-medium truncate">{fileName}</p>
              {fileSize && (
                <p className="text-xs text-muted-foreground">
                  {(fileSize / 1024 / 1024).toFixed(2)} MB
                </p>
              )}
            </div>
            <Button variant="ghost" size="icon" className="shrink-0">
              <Download className="h-4 w-4" />
            </Button>
          </div>
        );
      case 'system':
        return (
          <div className="text-center text-sm text-muted-foreground py-2">
            {message.content}
          </div>
        );
      default:
        return <p className="whitespace-pre-wrap break-words">{message.content}</p>;
    }
  };

  if (message.message_type === 'system') {
    return (
      <div className="flex justify-center py-2 animate-fade-in">
        <div className="px-4 py-1.5 rounded-full bg-muted/50 text-xs text-muted-foreground">
          {message.content}
        </div>
      </div>
    );
  }

  return (
    <div
      className={cn(
        'flex gap-3 group',
        isOwn ? 'flex-row-reverse' : 'flex-row',
        isOwn ? 'animate-slide-in-right' : 'animate-slide-in-left'
      )}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
    >
      {showAvatar && message.user ? (
        <UserAvatar user={message.user} size="sm" className="mt-1 shrink-0" />
      ) : (
        <div className="w-8 shrink-0" />
      )}

      <div className={cn('flex flex-col max-w-[70%]', isOwn ? 'items-end' : 'items-start')}>
        {/* Reply preview */}
        {message.reply_to_message && (
          <div
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 mb-1 rounded-lg text-xs',
              'bg-muted/50 border-l-2 border-primary/50'
            )}
          >
            <Reply className="h-3 w-3 text-muted-foreground" />
            <span className="font-medium text-muted-foreground">
              {message.reply_to_message.user?.username}
            </span>
            <span className="text-muted-foreground truncate max-w-[150px]">
              {message.reply_to_message.content}
            </span>
          </div>
        )}

        {/* Message bubble */}
        <div
          className={cn(
            'relative px-4 py-2.5 rounded-2xl',
            isOwn
              ? 'message-bubble-sent text-message-sent-foreground rounded-br-md'
              : 'message-bubble-received text-message-received-foreground rounded-bl-md'
          )}
        >
          {/* Sender name for group chats */}
          {!isOwn && message.user && showAvatar && (
            <p className="text-xs font-medium text-primary mb-1">{message.user.username}</p>
          )}

          {renderContent()}

          {/* Time and status */}
          <div className={cn('flex items-center gap-1.5 mt-1', isOwn ? 'justify-end' : 'justify-start')}>
            <span className="text-[10px] opacity-70">
              {format(new Date(message.created_at), 'HH:mm')}
            </span>
            {isEdited && <span className="text-[10px] opacity-50">(edited)</span>}
            {isOwn && message.status && <ReadReceipt status={message.status} className="ml-0.5" />}
          </div>
        </div>

        {/* Reactions */}
        {message.reactions && message.reactions.length > 0 && (
          <MessageReactions
            reactions={message.reactions}
            currentUserId={currentUserId}
            onAddReaction={(emoji) => onAddReaction?.(message.id, emoji)}
            onRemoveReaction={(id) => onRemoveReaction?.(id)}
            className="mt-1"
          />
        )}
      </div>

      {/* Actions */}
      <div
        className={cn(
          'flex items-center gap-1 opacity-0 transition-opacity self-center',
          showActions && 'opacity-100'
        )}
      >
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-7 w-7 rounded-full">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align={isOwn ? 'end' : 'start'} className="w-40">
            <DropdownMenuItem onClick={() => onReply?.(message)}>
              <Reply className="h-4 w-4 mr-2" />
              Reply
            </DropdownMenuItem>
            {isOwn && !isDeleted && (
              <>
                <DropdownMenuItem onClick={() => onEdit?.(message)}>
                  <Pencil className="h-4 w-4 mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => onDelete?.(message)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}

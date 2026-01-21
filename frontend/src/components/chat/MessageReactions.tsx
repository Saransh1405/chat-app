import { cn } from '@/lib/utils';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import type { MessageReaction } from '@/types/chat';

interface MessageReactionsProps {
  reactions: MessageReaction[];
  currentUserId: string;
  onAddReaction: (emoji: string) => void;
  onRemoveReaction: (reactionId: string) => void;
  className?: string;
}

const quickReactions = ['👍', '❤️', '😂', '😮', '😢', '🙏'];

export function MessageReactions({
  reactions,
  currentUserId,
  onAddReaction,
  onRemoveReaction,
  className,
}: MessageReactionsProps) {
  // Group reactions by emoji
  const groupedReactions = reactions.reduce((acc, reaction) => {
    if (!acc[reaction.reaction]) {
      acc[reaction.reaction] = [];
    }
    acc[reaction.reaction].push(reaction);
    return acc;
  }, {} as Record<string, MessageReaction[]>);

  const handleReactionClick = (emoji: string, reactions: MessageReaction[]) => {
    const userReaction = reactions.find((r) => r.user_id === currentUserId);
    if (userReaction) {
      onRemoveReaction(userReaction.id);
    } else {
      onAddReaction(emoji);
    }
  };

  return (
    <div className={cn('flex flex-wrap items-center gap-1', className)}>
      {Object.entries(groupedReactions).map(([emoji, reactionList]) => {
        const hasUserReacted = reactionList.some((r) => r.user_id === currentUserId);
        return (
          <button
            key={emoji}
            onClick={() => handleReactionClick(emoji, reactionList)}
            className={cn(
              'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs transition-colors',
              hasUserReacted
                ? 'bg-primary/20 text-primary border border-primary/30'
                : 'bg-muted hover:bg-muted/80 border border-transparent'
            )}
          >
            <span>{emoji}</span>
            <span className="font-medium">{reactionList.length}</span>
          </button>
        );
      })}
      
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="ghost" size="icon" className="h-6 w-6 rounded-full hover:bg-muted">
            <Plus className="h-3 w-3" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-2" side="top">
          <div className="flex gap-1">
            {quickReactions.map((emoji) => (
              <button
                key={emoji}
                onClick={() => onAddReaction(emoji)}
                className="p-2 text-lg hover:bg-muted rounded-md transition-colors"
              >
                {emoji}
              </button>
            ))}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}

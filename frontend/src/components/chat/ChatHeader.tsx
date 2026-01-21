import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { UserAvatar } from './UserAvatar';
import { Phone, Video, Info, MoreVertical, Users, Hash, User } from 'lucide-react';
import type { Room } from '@/types/chat';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface ChatHeaderProps {
  room: Room;
  onInfoClick?: () => void;
  membersCount?: number;
  isTyping?: boolean;
  typingUserName?: string;
}

const typeIcons = {
  group: Users,
  direct: User,
  channel: Hash,
};

export function ChatHeader({ room, onInfoClick, membersCount, isTyping, typingUserName }: ChatHeaderProps) {
  const Icon = typeIcons[room.type];

  return (
    <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-card/50 backdrop-blur-sm">
      <div className="flex items-center gap-3">
        <div className={cn(
          'flex items-center justify-center h-10 w-10 rounded-full',
          'bg-gradient-to-br from-primary/20 to-primary/5 text-primary'
        )}>
          <Icon className="h-5 w-5" />
        </div>
        <div>
          <h2 className="font-semibold">{room.name}</h2>
          <p className="text-xs text-muted-foreground">
            {isTyping && typingUserName ? (
              <span className="text-primary">{typingUserName} is typing...</span>
            ) : (
              <>
                {room.type === 'direct' && 'Online'}
                {room.type === 'group' && `${membersCount || 0} members`}
                {room.type === 'channel' && `${membersCount || 0} subscribers`}
              </>
            )}
          </p>
        </div>
      </div>

      <div className="flex items-center gap-1">
        <Button variant="ghost" size="icon" className="h-9 w-9 rounded-full hover:bg-muted">
          <Phone className="h-4 w-4 text-muted-foreground" />
        </Button>
        <Button variant="ghost" size="icon" className="h-9 w-9 rounded-full hover:bg-muted">
          <Video className="h-4 w-4 text-muted-foreground" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9 rounded-full hover:bg-muted"
          onClick={onInfoClick}
        >
          <Info className="h-4 w-4 text-muted-foreground" />
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-9 w-9 rounded-full hover:bg-muted">
              <MoreVertical className="h-4 w-4 text-muted-foreground" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuItem>Search messages</DropdownMenuItem>
            <DropdownMenuItem>Mute notifications</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="text-destructive focus:text-destructive">
              Leave {room.type === 'channel' ? 'channel' : 'group'}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}

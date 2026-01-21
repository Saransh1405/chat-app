import { cn } from '@/lib/utils';
import { UserAvatar } from './UserAvatar';
import { formatDistanceToNow } from 'date-fns';
import { Users, Hash, User } from 'lucide-react';
import type { Room } from '@/types/chat';

interface RoomListItemProps {
  room: Room;
  isActive?: boolean;
  onClick?: () => void;
}

const typeIcons = {
  group: Users,
  direct: User,
  channel: Hash,
};

export function RoomListItem({ room, isActive = false, onClick }: RoomListItemProps) {
  const Icon = typeIcons[room.type];
  const lastMessageTime = room.last_message?.created_at
    ? formatDistanceToNow(new Date(room.last_message.created_at), { addSuffix: false })
    : null;

  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full flex items-center gap-3 p-3 rounded-xl transition-all duration-200',
        'hover:bg-muted/50',
        isActive && 'bg-primary/10 hover:bg-primary/15'
      )}
    >
      {/* Avatar */}
      <div className="relative shrink-0">
        <div className={cn(
          'flex items-center justify-center h-12 w-12 rounded-full',
          'bg-gradient-to-br from-primary/20 to-primary/5 text-primary'
        )}>
          <Icon className="h-5 w-5" />
        </div>
        {room.unread_count && room.unread_count > 0 && (
          <span className="absolute -top-1 -right-1 flex items-center justify-center min-w-[20px] h-5 px-1.5 text-xs font-semibold bg-primary text-primary-foreground rounded-full">
            {room.unread_count > 99 ? '99+' : room.unread_count}
          </span>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0 text-left">
        <div className="flex items-center justify-between gap-2">
          <h4 className={cn(
            'font-medium truncate',
            room.unread_count && room.unread_count > 0 && 'text-foreground'
          )}>
            {room.name}
          </h4>
          {lastMessageTime && (
            <span className="text-xs text-muted-foreground shrink-0">
              {lastMessageTime}
            </span>
          )}
        </div>
        {room.last_message && (
          <p className={cn(
            'text-sm truncate mt-0.5',
            room.unread_count && room.unread_count > 0
              ? 'text-foreground font-medium'
              : 'text-muted-foreground'
          )}>
            {room.last_message.user?.username && (
              <span className="text-muted-foreground">{room.last_message.user.username}: </span>
            )}
            {room.last_message.deleted_at
              ? 'Message deleted'
              : room.last_message.content}
          </p>
        )}
      </div>
    </button>
  );
}

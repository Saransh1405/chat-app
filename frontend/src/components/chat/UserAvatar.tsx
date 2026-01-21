import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';

interface UserAvatarProps {
  user: {
    username: string;
    avatar_url?: string;
  };
  size?: 'sm' | 'md' | 'lg';
  showOnlineStatus?: boolean;
  isOnline?: boolean;
  className?: string;
}

const sizeClasses = {
  sm: 'h-8 w-8',
  md: 'h-10 w-10',
  lg: 'h-12 w-12',
};

const statusSizeClasses = {
  sm: 'h-2.5 w-2.5 border-[1.5px]',
  md: 'h-3 w-3 border-2',
  lg: 'h-3.5 w-3.5 border-2',
};

export function UserAvatar({ user, size = 'md', showOnlineStatus = false, isOnline = false, className }: UserAvatarProps) {
  const initials = user.username
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  return (
    <div className={cn('relative', className)}>
      <Avatar className={cn(sizeClasses[size], 'ring-2 ring-background')}>
        <AvatarImage src={user.avatar_url} alt={user.username} />
        <AvatarFallback className="bg-gradient-to-br from-primary/80 to-primary text-primary-foreground font-medium text-xs">
          {initials}
        </AvatarFallback>
      </Avatar>
      {showOnlineStatus && (
        <span
          className={cn(
            'absolute bottom-0 right-0 rounded-full border-background',
            statusSizeClasses[size],
            isOnline ? 'bg-online' : 'bg-offline'
          )}
        />
      )}
    </div>
  );
}

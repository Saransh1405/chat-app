import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { RoomListItem } from './RoomListItem';
import { UserAvatar } from './UserAvatar';
import { useChatStore } from '@/store/chatStore';
import { Search, Plus, Settings, LogOut, MessageSquare } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface ChatSidebarProps {
  onCreateRoom?: () => void;
}

export function ChatSidebar({ onCreateRoom }: ChatSidebarProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const { rooms, currentRoom, setCurrentRoom, currentUser, logout } = useChatStore();

  const filteredRooms = rooms.filter((room) =>
    room.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="flex flex-col h-full bg-sidebar border-r border-sidebar-border">
      {/* Header */}
      <div className="p-4 border-b border-sidebar-border">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="flex items-center justify-center h-9 w-9 rounded-lg bg-primary">
              <MessageSquare className="h-5 w-5 text-primary-foreground" />
            </div>
            <h1 className="text-lg font-semibold">ChatSDK</h1>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="h-9 w-9 rounded-full hover:bg-sidebar-accent"
            onClick={onCreateRoom}
          >
            <Plus className="h-5 w-5" />
          </Button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search conversations..."
            className="pl-9 bg-sidebar-accent border-0 focus-visible:ring-1 focus-visible:ring-sidebar-ring"
          />
        </div>
      </div>

      {/* Room list */}
      <ScrollArea className="flex-1 px-2">
        <div className="py-2 space-y-1">
          {filteredRooms.length > 0 ? (
            filteredRooms.map((room) => (
              <RoomListItem
                key={room.id}
                room={room}
                isActive={currentRoom?.id === room.id}
                onClick={() => setCurrentRoom(room)}
              />
            ))
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="p-4 rounded-full bg-muted/50 mb-4">
                <MessageSquare className="h-8 w-8 text-muted-foreground" />
              </div>
              <p className="text-sm text-muted-foreground">
                {searchQuery ? 'No conversations found' : 'No conversations yet'}
              </p>
              {!searchQuery && (
                <Button variant="link" size="sm" onClick={onCreateRoom} className="mt-2">
                  Start a new conversation
                </Button>
              )}
            </div>
          )}
        </div>
      </ScrollArea>

      {/* User section */}
      {currentUser && (
        <div className="p-3 border-t border-sidebar-border">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="flex items-center gap-3 w-full p-2 rounded-lg hover:bg-sidebar-accent transition-colors">
                <UserAvatar user={currentUser} size="sm" showOnlineStatus isOnline />
                <div className="flex-1 text-left min-w-0">
                  <p className="font-medium text-sm truncate">{currentUser.username}</p>
                  <p className="text-xs text-muted-foreground truncate">{currentUser.email}</p>
                </div>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-56">
              <DropdownMenuItem>
                <Settings className="h-4 w-4 mr-2" />
                Settings
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={logout} className="text-destructive focus:text-destructive">
                <LogOut className="h-4 w-4 mr-2" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}
    </div>
  );
}

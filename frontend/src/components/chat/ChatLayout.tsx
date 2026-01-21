import { useState } from 'react';
import { ChatSidebar } from './ChatSidebar';
import { ChatArea } from './ChatArea';
import { CreateRoomDialog } from './CreateRoomDialog';
import { useChatStore } from '@/store/chatStore';

export function ChatLayout() {
  const [createRoomOpen, setCreateRoomOpen] = useState(false);
  const { currentRoom } = useChatStore();

  return (
    <div className="flex h-screen w-full overflow-hidden">
      {/* Sidebar */}
      <div className="w-80 shrink-0">
        <ChatSidebar onCreateRoom={() => setCreateRoomOpen(true)} />
      </div>

      {/* Main chat area */}
      <ChatArea room={currentRoom} />

      {/* Create room dialog */}
      <CreateRoomDialog open={createRoomOpen} onOpenChange={setCreateRoomOpen} />
    </div>
  );
}

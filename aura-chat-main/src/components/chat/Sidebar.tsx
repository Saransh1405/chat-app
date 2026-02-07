import { motion } from "framer-motion";
import { Hash, Lock, MessageSquare, Menu, LogOut, Search } from "lucide-react";
import { cn } from "@/lib/utils";
import { ThemeToggle } from "./ThemeToggle";
import { UserAvatar } from "./UserAvatar";
import { CreateRoomDialog } from "./CreateRoomDialog";
import type { Room, User } from "@/types/chat";
import { useState } from "react";

interface SidebarProps {
  rooms: Room[];
  activeRoomId: string | null;
  currentUser: User | null;
  onSelectRoom: (roomId: string) => void;
  onCreateRoom: (name: string, type: Room["type"]) => void;
  onLogout: () => void;
  isOpen: boolean;
  onToggle: () => void;
}

const ROOM_GRADIENTS = [
  "from-purple-500 to-pink-500",
  "from-blue-500 to-cyan-500",
  "from-emerald-500 to-teal-500",
  "from-orange-500 to-amber-500",
  "from-rose-500 to-pink-400",
];

function getRoomGradient(index: number) {
  return ROOM_GRADIENTS[index % ROOM_GRADIENTS.length];
}

export function Sidebar({
  rooms,
  activeRoomId,
  currentUser,
  onSelectRoom,
  onCreateRoom,
  onLogout,
  isOpen,
  onToggle,
}: SidebarProps) {
  const [search, setSearch] = useState("");

  const filteredRooms = rooms.filter((r) =>
    r.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm md:hidden"
          onClick={onToggle}
        />
      )}

      {/* Mobile toggle */}
      <button
        onClick={onToggle}
        className="fixed left-4 top-4 z-50 flex h-10 w-10 items-center justify-center rounded-xl glass md:hidden"
      >
        <Menu className="h-5 w-5" />
      </button>

      {/* Sidebar */}
      <motion.aside
        initial={{ x: -320 }}
        animate={{ x: isOpen ? 0 : (window.innerWidth >= 768 ? 0 : -320) }}
        transition={{ type: "spring", damping: 25, stiffness: 300 }}
        className={cn(
          "fixed left-0 top-0 z-50 flex h-full w-80 flex-col border-r border-border glass-strong",
          "md:sticky md:top-0 md:translate-x-0"
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 pb-2">
          <h1 className="gradient-text text-2xl font-extrabold tracking-tight">Aura</h1>
          <div className="flex items-center gap-1">
            <ThemeToggle />
            <CreateRoomDialog onCreateRoom={onCreateRoom} />
          </div>
        </div>

        {/* Search */}
        <div className="px-4 py-2">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search rooms..."
              className="w-full rounded-xl bg-muted/50 py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/30"
            />
          </div>
        </div>

        {/* Room list */}
        <nav className="flex-1 overflow-y-auto scrollbar-thin px-3 py-2 space-y-0.5">
          {filteredRooms.map((room, i) => {
            const isActive = room.id === activeRoomId;
            return (
              <button
                key={room.id}
                onClick={() => {
                  onSelectRoom(room.id);
                  if (window.innerWidth < 768) onToggle();
                }}
                className={cn(
                  "group flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left transition-all",
                  isActive
                    ? "bg-primary/10 shadow-sm"
                    : "hover:bg-muted/50"
                )}
              >
                {/* Room avatar */}
                <div
                  className={cn(
                    "flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br text-white text-sm font-bold shadow-md transition-shadow",
                    getRoomGradient(i),
                    isActive && "glow-primary"
                  )}
                >
                  {room.type === "private" ? (
                    <Lock className="h-4 w-4" />
                  ) : (
                    <Hash className="h-4 w-4" />
                  )}
                </div>

                {/* Room info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <span
                      className={cn(
                        "truncate text-sm font-medium",
                        isActive ? "text-primary" : "text-foreground"
                      )}
                    >
                      {room.name}
                    </span>
                    {(room.unread_count ?? 0) > 0 && (
                      <span className="flex h-5 min-w-[20px] items-center justify-center rounded-full gradient-primary px-1.5 text-[10px] font-bold text-white">
                        {room.unread_count}
                      </span>
                    )}
                  </div>
                  {room.last_message && (
                    <p className="truncate text-xs text-muted-foreground mt-0.5">
                      {room.last_message}
                    </p>
                  )}
                </div>
              </button>
            );
          })}
        </nav>

        {/* User profile bar */}
        {currentUser && (
          <div className="border-t border-border p-3">
            <div className="flex items-center gap-3 rounded-xl p-2">
              <UserAvatar
                userId={currentUser.id}
                username={currentUser.username}
                avatarUrl={currentUser.avatar_url}
                size="md"
                isOnline
              />
              <div className="flex-1 min-w-0">
                <p className="truncate text-sm font-medium">{currentUser.username}</p>
                <p className="truncate text-xs text-muted-foreground">{currentUser.email}</p>
              </div>
              <button
                onClick={onLogout}
                className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                title="Log out"
              >
                <LogOut className="h-4 w-4" />
              </button>
            </div>
          </div>
        )}
      </motion.aside>
    </>
  );
}

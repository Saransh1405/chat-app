import type { Room } from "@/types/chat";
import { InviteDialog } from "./InviteDialog";
import { Hash, Lock, Users, Book, PenSquare } from "lucide-react";
import { cn } from "@/lib/utils";

/** Base URL of the whiteboard (canvas) app. Same room ID = same whiteboard for all chat members. */
const WHITEBOARD_APP_URL = import.meta.env.VITE_WHITEBOARD_APP_URL || "http://localhost:5000";

interface RoomHeaderProps {
    room: Room | null;
    isAdmin?: boolean;
    isMember?: boolean;
    onJoin?: () => void;
    mode?: "chat" | "knowledge_base";
}

export function RoomHeader({ room, isAdmin = true, isMember = true, onJoin, mode = "chat" }: RoomHeaderProps) {
    if (mode === "knowledge_base") {
        return (
            <header className="flex h-16 items-center justify-between border-b border-border px-6">
                <div className="flex items-center gap-3">
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                        <Book className="h-4 w-4" />
                    </div>
                    <div>
                        <h2 className="text-sm font-semibold">AI Knowledge Assistant</h2>
                        <div className="flex items-center gap-1 text-xs text-muted-foreground">
                            <span>Ask questions about your documents</span>
                        </div>
                    </div>
                </div>
            </header>
        )
    }

    if (!room) {
        return (
            <header className="flex h-16 items-center border-b border-border px-6">
                <p className="text-sm text-muted-foreground">Select a room to start chatting</p>
            </header>
        );
    }

    return (
        <header className="flex h-16 items-center justify-between border-b border-border px-6">
            <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
                    {room.type === "private" ? (
                        <Lock className="h-4 w-4 text-muted-foreground" />
                    ) : (
                        <Hash className="h-4 w-4 text-muted-foreground" />
                    )}
                </div>
                <div>
                    <h2 className="text-sm font-semibold">{room.name}</h2>
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                        <Users className="h-3 w-3" />
                        <span>{room.member_count} members</span>
                    </div>
                </div>
            </div>

            <div className="flex items-center gap-2">
                <a
                    href={`${WHITEBOARD_APP_URL}/room/${room.id}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                    title="Open collaborative whiteboard (same room)"
                >
                    <PenSquare className="h-4 w-4" />
                    <span>Whiteboard</span>
                </a>
                {!isMember && (
                    <button
                        onClick={onJoin}
                        className="rounded-xl gradient-primary px-4 py-1.5 text-xs font-semibold text-white shadow-lg transition-all hover:shadow-xl"
                    >
                        Join Room
                    </button>
                )}
                {isAdmin && isMember && (
                    <InviteDialog roomId={room.id} roomName={room.name} />
                )}
            </div>
        </header>
    );
}

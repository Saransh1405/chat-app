import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { UserPlus, Search } from "lucide-react";
import { roomsAPI } from "@/services/roomsAPI";
import { usersAPI } from "@/services/usersAPI";
import type { User } from "@/types/chat";
import { UserAvatar } from "./UserAvatar";
import { toast } from "sonner";

interface InviteDialogProps {
  roomId: string;
  roomName: string;
}

export function InviteDialog({ roomId, roomName }: InviteDialogProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const handleSearch = async (value: string) => {
    setQuery(value);
    if (value.length < 2) {
      setResults([]);
      return;
    }
    setIsSearching(true);
    try {
      const users = await usersAPI.search(value);
      setResults(users);
    } catch {
      setResults([]);
    }
    setIsSearching(false);
  };

  const handleInvite = async (user: User) => {
    try {
      await roomsAPI.invite(roomId, user.email);
      toast.success(`Invited ${user.username} to ${roomName}`);
    } catch {
      toast.error("Failed to send invite");
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
          <UserPlus className="h-4 w-4" />
          <span>Invite</span>
        </button>
      </DialogTrigger>
      <DialogContent className="glass-strong sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="gradient-text text-xl font-bold">Invite to {roomName}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 pt-2">
          <div className="space-y-2">
            <Label>Search users</Label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(e) => handleSearch(e.target.value)}
                placeholder="Search by username or email..."
                className="bg-muted/50 pl-9"
                autoFocus
              />
            </div>
          </div>

          {/* Results */}
          <div className="max-h-60 space-y-1 overflow-y-auto scrollbar-thin">
            {results.map((user) => (
              <div
                key={user.id}
                className="flex items-center gap-3 rounded-lg p-2 transition-colors hover:bg-muted/50"
              >
                <UserAvatar userId={user.id} username={user.username} avatarUrl={user.avatar_url} size="sm" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{user.username}</p>
                  <p className="text-xs text-muted-foreground truncate">{user.email}</p>
                </div>
                <button
                  onClick={() => handleInvite(user)}
                  className="rounded-lg gradient-primary px-3 py-1 text-xs font-medium text-white shadow-sm transition-all hover:shadow-md"
                >
                  Invite
                </button>
              </div>
            ))}
            {query.length >= 2 && results.length === 0 && !isSearching && (
              <p className="py-4 text-center text-sm text-muted-foreground">No users found</p>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

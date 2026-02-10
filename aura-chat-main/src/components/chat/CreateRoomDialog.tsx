import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Plus } from "lucide-react";
import type { Room } from "@/types/chat";
import { cn } from "@/lib/utils";

interface CreateRoomDialogProps {
  onCreateRoom: (name: string, type: Room["type"]) => void;
}

export function CreateRoomDialog({ onCreateRoom }: CreateRoomDialogProps) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [type, setType] = useState<Room["type"]>("public");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    onCreateRoom(name.trim(), type);
    setName("");
    setType("public");
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
          <Plus className="h-5 w-5" />
        </button>
      </DialogTrigger>
      <DialogContent className="glass-strong sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="gradient-text text-xl font-bold">Create Room</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-2">
            <Label htmlFor="room-name">Room Name</Label>
            <Input
              id="room-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Design Team"
              className="bg-muted/50"
              autoFocus
            />
          </div>
          <div className="space-y-2">
            <Label>Type</Label>
            <div className="flex gap-2">
              {(["public", "private"] as const).map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setType(t)}
                  className={cn(
                    "flex-1 rounded-lg border px-4 py-2 text-sm font-medium capitalize transition-all",
                    type === t
                      ? "gradient-primary border-primary/50 text-white shadow-md"
                      : "border-border bg-muted/50 text-muted-foreground hover:border-primary/30"
                  )}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>
          <button
            type="submit"
            disabled={!name.trim()}
            className="w-full rounded-xl gradient-primary py-2.5 text-sm font-semibold text-white shadow-lg transition-all hover:shadow-xl disabled:opacity-50"
          >
            Create Room
          </button>
        </form>
      </DialogContent>
    </Dialog>
  );
}

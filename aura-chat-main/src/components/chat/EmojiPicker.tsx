import { useState, useRef } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Smile } from "lucide-react";
import { cn } from "@/lib/utils";

const EMOJI_LIST = [
  "👍", "❤️", "😂", "😮", "😢", "🔥",
  "🚀", "✨", "💜", "🎉", "✅", "👀",
  "💪", "🙌", "🤔", "👏", "💯", "⭐",
];

interface EmojiPickerProps {
  onSelect: (emoji: string) => void;
  trigger?: React.ReactNode;
  className?: string;
}

export function EmojiPicker({ onSelect, trigger, className }: EmojiPickerProps) {
  const [open, setOpen] = useState(false);

  const handleSelect = (emoji: string) => {
    onSelect(emoji);
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        {trigger || (
          <button
            className={cn(
              "flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground",
              className
            )}
          >
            <Smile className="h-5 w-5" />
          </button>
        )}
      </PopoverTrigger>
      <PopoverContent
        className="w-auto p-2 glass"
        align="end"
        side="top"
        sideOffset={8}
      >
        <div className="grid grid-cols-6 gap-1">
          {EMOJI_LIST.map((emoji) => (
            <button
              key={emoji}
              onClick={() => handleSelect(emoji)}
              className="flex h-9 w-9 items-center justify-center rounded-md text-lg transition-transform hover:scale-125 hover:bg-muted"
            >
              {emoji}
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}

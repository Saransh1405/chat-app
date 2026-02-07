import { useState, useRef, useCallback } from "react";
import { Send, Paperclip, X } from "lucide-react";
import { EmojiPicker } from "./EmojiPicker";
import { cn } from "@/lib/utils";

interface MessageInputProps {
  onSend: (content: string, imageUrl?: string) => void;
  onTyping: () => void;
  disabled?: boolean;
}

export function MessageInput({ onSend, onTyping, disabled }: MessageInputProps) {
  const [content, setContent] = useState("");
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [imageName, setImageName] = useState<string>("");
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textInputRef = useRef<HTMLInputElement>(null);

  const handleSend = useCallback(() => {
    const text = content.trim();
    if (!text && !imagePreview) return;
    onSend(text, imagePreview || undefined);
    setContent("");
    setImagePreview(null);
    setImageName("");
    textInputRef.current?.focus();
  }, [content, imagePreview, onSend]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setContent(e.target.value);
    onTyping();
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setImageName(file.name);
    const url = URL.createObjectURL(file);
    setImagePreview(url);
  };

  const handleEmojiSelect = (emoji: string) => {
    setContent((prev) => prev + emoji);
    textInputRef.current?.focus();
  };

  return (
    <div className="p-4">
      {/* Image preview */}
      {imagePreview && (
        <div className="mb-2 flex items-center gap-2 rounded-lg bg-muted/50 p-2">
          <img src={imagePreview} alt="Preview" className="h-16 w-16 rounded-md object-cover" />
          <div className="flex-1 min-w-0">
            <p className="truncate text-xs text-muted-foreground">{imageName}</p>
          </div>
          <button
            onClick={() => {
              setImagePreview(null);
              setImageName("");
            }}
            className="flex h-6 w-6 items-center justify-center rounded-full hover:bg-muted"
          >
            <X className="h-4 w-4 text-muted-foreground" />
          </button>
        </div>
      )}

      {/* Input bar */}
      <div className="glass flex items-center gap-2 rounded-2xl px-3 py-2">
        {/* File attach */}
        <button
          onClick={() => fileInputRef.current?.click()}
          className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <Paperclip className="h-5 w-5" />
        </button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileSelect}
          className="hidden"
        />

        {/* Text input */}
        <input
          ref={textInputRef}
          type="text"
          value={content}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder="Type a message..."
          disabled={disabled}
          className="flex-1 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none disabled:opacity-50"
        />

        {/* Emoji */}
        <EmojiPicker onSelect={handleEmojiSelect} />

        {/* Send */}
        <button
          onClick={handleSend}
          disabled={(!content.trim() && !imagePreview) || disabled}
          className={cn(
            "flex h-9 w-9 items-center justify-center rounded-xl transition-all",
            content.trim() || imagePreview
              ? "gradient-primary text-white shadow-lg glow-primary"
              : "bg-muted text-muted-foreground"
          )}
        >
          <Send className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}

import { useState, useRef, useCallback } from "react";
import { Send, Paperclip, X, File, Loader2 } from "lucide-react";
import { EmojiPicker } from "./EmojiPicker";
import { cn } from "@/lib/utils";
import { fileUploadAPI } from "@/services/fileUploadAPI";
import { useAuth } from "@/contexts/AuthContext";

interface MessageInputProps {
  onSend: (content: string, imageUrl?: string, fileData?: {
    filename: string;
    file_path: string;
    file_size: number;
    mime_type: string;
  }) => void;
  onTyping: () => void;
  disabled?: boolean;
}

export function MessageInput({ onSend, onTyping, disabled }: MessageInputProps) {
  const { user } = useAuth();
  const [content, setContent] = useState("");
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [imageName, setImageName] = useState<string>("");
  const [imageFile, setImageFile] = useState<File | null>(null); // Store the image file
  const [fileData, setFileData] = useState<{
    file: File;
    preview: string;
    data?: {
      filename: string;
      file_path: string;
      file_size: number;
      mime_type: string;
    };
  } | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textInputRef = useRef<HTMLInputElement>(null);

  const handleSend = useCallback(async () => {
    const text = content.trim();
    if (!text && !imagePreview && !fileData) return;

      setIsUploading(true);
    try {
      // If there's an image preview (image file selected), upload it first
      if (imagePreview && imageName && imageFile) {
        try {
          const uploadResponse = await fileUploadAPI.upload(
            imageFile,
            user?.application_id
          );
          
          const uploadedFileData = {
            filename: uploadResponse.data.filename,
            file_path: uploadResponse.data.file_path,
            file_size: uploadResponse.data.size,
            mime_type: uploadResponse.data.mime_type,
          };

          // Send message with image file data as MEDIA
          onSend(text, undefined, uploadedFileData);
          
          // Reset state
          setContent("");
          setFileData(null);
          setImagePreview(null);
          setImageName("");
          setImageFile(null);
          if (fileInputRef.current) fileInputRef.current.value = "";
          textInputRef.current?.focus();
          return;
        } catch (error) {
          console.error("[MessageInput] Failed to upload image:", error);
          alert("Failed to upload image. Please try again.");
          setIsUploading(false);
          return;
        }
      }

      // If there's a non-image file, upload it first
      if (fileData && !fileData.data) {
        try {
          const uploadResponse = await fileUploadAPI.upload(
            fileData.file,
            user?.application_id
          );
          
          const uploadedFileData = {
            filename: uploadResponse.data.filename,
            file_path: uploadResponse.data.file_path,
            file_size: uploadResponse.data.size,
            mime_type: uploadResponse.data.mime_type,
          };

          // Send message with file data as MEDIA
          onSend(text, undefined, uploadedFileData);
          
          // Reset state
          setContent("");
          setFileData(null);
          setImagePreview(null);
          setImageName("");
          if (fileInputRef.current) fileInputRef.current.value = "";
          textInputRef.current?.focus();
          return;
        } catch (error) {
          console.error("[MessageInput] Failed to upload file:", error);
          alert("Failed to upload file. Please try again.");
          setIsUploading(false);
          return;
        }
      }

      // If file data already exists (already uploaded), send it
      if (fileData?.data) {
        onSend(text, undefined, fileData.data);
      } else {
        // Send as TEXT message (no file)
        onSend(text, undefined, undefined);
      }
      
      // Reset state
      setContent("");
      setImagePreview(null);
      setImageName("");
      setImageFile(null);
      setFileData(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
      textInputRef.current?.focus();
    } finally {
      setIsUploading(false);
    }
  }, [content, imagePreview, imageName, imageFile, fileData, onSend, user?.application_id]);

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

    // Check if it's an image
    if (file.type.startsWith("image/")) {
      setImageName(file.name);
      const url = URL.createObjectURL(file);
      setImagePreview(url);
      setImageFile(file); // Store the file for upload
      setFileData(null); // Clear file data if image
    } else {
      // Handle other file types
      setFileData({
        file,
        preview: file.name,
      });
      setImagePreview(null); // Clear image preview
      setImageName("");
      setImageFile(null);
    }

    // Reset input to allow selecting the same file again if needed
    // The file is stored in state (imageFile or fileData.file)
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
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
              setImageFile(null);
            }}
            className="flex h-6 w-6 items-center justify-center rounded-full hover:bg-muted"
            disabled={isUploading}
          >
            <X className="h-4 w-4 text-muted-foreground" />
          </button>
        </div>
      )}

      {/* File preview */}
      {fileData && !imagePreview && (
        <div className="mb-2 flex items-center gap-2 rounded-lg bg-muted/50 p-2">
          <div className="flex h-16 w-16 items-center justify-center rounded-md bg-primary/10">
            <File className="h-8 w-8 text-primary" />
          </div>
          <div className="flex-1 min-w-0">
            <p className="truncate text-xs font-medium text-foreground">{fileData.file.name}</p>
            <p className="truncate text-xs text-muted-foreground">
              {(fileData.file.size / 1024).toFixed(1)} KB
            </p>
          </div>
          <button
            onClick={() => {
              setFileData(null);
            }}
            className="flex h-6 w-6 items-center justify-center rounded-full hover:bg-muted"
            disabled={isUploading}
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
          accept="*/*"
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
          disabled={(!content.trim() && !imagePreview && !fileData) || disabled || isUploading}
          className={cn(
            "flex h-9 w-9 items-center justify-center rounded-xl transition-all",
            (content.trim() || imagePreview || fileData) && !isUploading
              ? "gradient-primary text-white shadow-lg glow-primary"
              : "bg-muted text-muted-foreground"
          )}
        >
          {isUploading ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Send className="h-4 w-4" />
          )}
        </button>
      </div>
    </div>
  );
}

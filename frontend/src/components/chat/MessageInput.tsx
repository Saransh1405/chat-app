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
    mode?: "chat" | "knowledge_base";
}

export function MessageInput({ onSend, onTyping, disabled, mode = "chat" }: MessageInputProps) {
    const { user } = useAuth();
    const [content, setContent] = useState("");
    const [imagePreview, setImagePreview] = useState<string | null>(null);
    const [imageName, setImageName] = useState<string>("");
    const [imageFile, setImageFile] = useState<File | null>(null);
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

                    onSend(text, undefined, uploadedFileData);

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

                    onSend(text, undefined, uploadedFileData);

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

            if (fileData?.data) {
                onSend(text, undefined, fileData.data);
            } else {
                onSend(text, undefined, undefined);
            }

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

        if (file.type.startsWith("image/")) {
            setImageName(file.name);
            const url = URL.createObjectURL(file);
            setImagePreview(url);
            setImageFile(file);
            setFileData(null);
        } else {
            setFileData({
                file,
                preview: file.name,
            });
            setImagePreview(null);
            setImageName("");
            setImageFile(null);
        }

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

            <div className="flex items-center gap-2 rounded-2xl border border-border/60 bg-card/50 px-3 py-2 shadow-sm backdrop-blur-sm">
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

                <input
                    ref={textInputRef}
                    type="text"
                    value={content}
                    onChange={handleChange}
                    onKeyDown={handleKeyDown}
                    placeholder={mode === "knowledge_base" ? "Ask about your documents..." : "Type a message..."}
                    disabled={disabled}
                    className="flex-1 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none disabled:opacity-50"
                />

                <EmojiPicker onSelect={handleEmojiSelect} />

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

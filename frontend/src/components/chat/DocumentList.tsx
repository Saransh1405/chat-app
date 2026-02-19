import { useState, useRef, useEffect } from "react";
import { Upload, FileText, Trash2, HardDrive, CheckCircle, Loader2 } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { knowledgeBaseAPI, type Document } from "@/services/knowledgeBaseAPI";
import { cn } from "@/lib/utils";

interface DocumentListProps {
    className?: string;
}

export function DocumentList({ className }: DocumentListProps) {
    const [documents, setDocuments] = useState<Document[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isUploading, setIsUploading] = useState(false);
    const [dragActive, setDragActive] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        fetchDocuments();
    }, []);

    const fetchDocuments = async () => {
        try {
            setIsLoading(true);
            const docs = await knowledgeBaseAPI.list();
            setDocuments(docs);
        } catch (error) {
            console.error("Failed to fetch documents:", error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleDrag = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        if (e.type === "dragenter" || e.type === "dragover") {
            setDragActive(true);
        } else if (e.type === "dragleave") {
            setDragActive(false);
        }
    };

    const handleDrop = async (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setDragActive(false);

        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            await handleUpload(e.dataTransfer.files[0]);
        }
    };

    const handleUpload = async (file: File) => {
        try {
            setIsUploading(true);
            // Upload the file
            const response = await knowledgeBaseAPI.upload(file);
            console.log("File uploaded successfully:", response);
            
            // Call get files API after successful upload
            // Add a small delay to ensure the backend has processed the file
            await new Promise(resolve => setTimeout(resolve, 500));
            await fetchDocuments();
            console.log("Files list refreshed after upload");
        } catch (error) {
            console.error("Failed to upload document:", error);
            // Still try to refresh the list in case the file was partially processed
            try {
                await fetchDocuments();
            } catch (fetchError) {
                console.error("Failed to fetch documents after upload error:", fetchError);
            }
        } finally {
            setIsUploading(false);
        }
    };

    const handleDelete = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            setDocuments(prev => prev.filter(doc => doc.id !== id));
            await knowledgeBaseAPI.delete(id);
        } catch (error) {
            console.error("Failed to delete document:", error);
            fetchDocuments();
        }
    };

    const formatSize = (bytes: number) => {
        if (bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
    };

    const totalSize = documents.reduce((acc, doc) => acc + doc.file_size, 0);

    return (
        <div className={cn("flex flex-col h-full bg-background/50", className)}>
            <div className="p-4 pb-2">
                <div
                    className={cn(
                        "relative flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-muted-foreground/25 p-6 transition-colors",
                        dragActive ? "border-primary bg-primary/5" : "hover:border-primary/50 hover:bg-muted/50",
                        isUploading && "pointer-events-none opacity-50"
                    )}
                    onDragEnter={handleDrag}
                    onDragLeave={handleDrag}
                    onDragOver={handleDrag}
                    onDrop={handleDrop}
                    onClick={() => fileInputRef.current?.click()}
                >
                    <input
                        ref={fileInputRef}
                        type="file"
                        className="hidden"
                        accept=".pdf,.docx,.txt,.md"
                        onChange={(e) => {
                            if (e.target.files && e.target.files[0]) {
                                handleUpload(e.target.files[0]);
                            }
                        }}
                    />

                    {isUploading ? (
                        <div className="flex flex-col items-center gap-2">
                            <Loader2 className="h-8 w-8 animate-spin text-primary" />
                            <p className="text-sm font-medium">Uploading...</p>
                        </div>
                    ) : (
                        <>
                            <Upload className="mb-2 h-8 w-8 text-muted-foreground" />
                            <p className="text-sm font-medium text-foreground">Click to upload</p>
                            <p className="text-xs text-muted-foreground">or drag and drop PDF, DOCX, TXT</p>
                        </>
                    )}
                </div>
            </div>

            <div className="px-4 py-2 flex items-center justify-between text-xs text-muted-foreground">
                <span>{documents.length} Documents</span>
                <span>{formatSize(totalSize)} Used</span>
            </div>

            <div className="flex-1 overflow-y-auto px-2 space-y-2">
                {isLoading ? (
                    <div className="flex justify-center py-8">
                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                    </div>
                ) : documents.length === 0 ? (
                    <div className="flex flex-col items-center justify-center py-12 text-center text-muted-foreground">
                        <HardDrive className="h-12 w-12 mb-3 opacity-20" />
                        <p>No documents yet</p>
                    </div>
                ) : (
                    <AnimatePresence initial={false}>
                        {documents.map((doc) => (
                            <motion.div
                                key={doc.id}
                                initial={{ opacity: 0, y: 10 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, height: 0, overflow: "hidden" }}
                                className="group relative flex items-center gap-3 rounded-lg border border-border bg-card p-3 shadow-sm transition-all hover:bg-accent hover:shadow-md"
                            >
                                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                                    <FileText className="h-5 w-5" />
                                </div>

                                <div className="flex-1 min-w-0">
                                    <h4 className="truncate text-sm font-medium text-foreground" title={doc.file_name}>
                                        {doc.file_name}
                                    </h4>
                                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                        <span>{formatSize(doc.file_size)}</span>
                                        <span>•</span>
                                        <span className="flex items-center gap-1">
                                            <CheckCircle className="h-3 w-3 text-emerald-500" />
                                            Ready
                                        </span>
                                    </div>
                                </div>

                                <button
                                    onClick={(e) => handleDelete(doc.id, e)}
                                    className="absolute right-2 top-2 p-1.5 rounded-md opacity-0 text-muted-foreground hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100 transition-all"
                                    title="Delete document"
                                >
                                    <Trash2 className="h-4 w-4" />
                                </button>
                            </motion.div>
                        ))}
                    </AnimatePresence>
                )}
            </div>
        </div>
    );
}

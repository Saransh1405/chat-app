import { api } from "./api";

export interface Document {
    id: string;
    user_id: string;
    file_name: string;
    file_type: string;
    file_size: number;
    uploaded_at: string;
}

export const knowledgeBaseAPI = {
    async list(): Promise<Document[]> {
        const res = await api.get<{ documents: Document[] }>(`/upload`);
        return res.documents || [];
    },

    async upload(file: File): Promise<void> {
        await api.upload<{ message: string }>("/upload", file, {
            type: file.type,
            title: file.name,
            file_size: file.size.toString(),
        });
    },

    async delete(id: string): Promise<void> {
        await api.delete(`/upload?id=${id}`);
    },
};

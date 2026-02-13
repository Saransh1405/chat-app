import { api } from "./api";

export interface FileUploadResponse {
  message: string;
  data: {
    file_id: string;
    url: string;
    filename: string;
    size: number;
    mime_type: string;
    file_path: string;
    application_id?: string;
  };
}

export const fileUploadAPI = {
  async upload(file: File, applicationId?: string): Promise<FileUploadResponse> {
    const endpoint = applicationId 
      ? `/file/upload?application_id=${applicationId}` 
      : "/file/upload";
    
    // Use the upload method from api client
    const response = await api.upload<FileUploadResponse>(endpoint, file);
    return response;
  },
};


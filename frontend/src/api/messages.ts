import { apiClient } from './client';
import { Message } from '../types';

export const messagesAPI = {
  create: async (data: {
    room_id: string;
    content: string;
    application_id?: string;
    message_type?: 'text' | 'image' | 'file' | 'system';
    reply_to?: string;
    metadata?: Record<string, any>;
  }): Promise<{ message: Message; status: string }> => {
    const response = await apiClient.post('/messages', data);
    return response.data;
  },

  get: async (id: string, room_id: string, application_id?: string): Promise<{ message: Message }> => {
    const params: any = { id, room_id };
    if (application_id) params.application_id = application_id;
    const response = await apiClient.get('/messages', { params });
    return response.data;
  },

  list: async (room_id: string, application_id?: string, limit = 50, offset = 0): Promise<{
    messages: Message[];
    limit: number;
    offset: number;
    count: number;
  }> => {
    const params: any = { room_id, limit, offset };
    if (application_id) params.application_id = application_id;
    const response = await apiClient.get('/messages', { params });
    return response.data;
  },

  update: async (data: {
    id: string;
    room_id: string;
    application_id?: string;
    content?: string;
    message_type?: 'text' | 'image' | 'file' | 'system';
    metadata?: Record<string, any>;
  }): Promise<{ message: string; data: Message }> => {
    const response = await apiClient.patch('/messages', data);
    return response.data;
  },

  delete: async (id: string, room_id: string, application_id?: string): Promise<{ message: string }> => {
    const response = await apiClient.delete('/messages', { data: { id, room_id, application_id } });
    return response.data;
  },
};


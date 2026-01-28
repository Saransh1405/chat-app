import { apiClient } from './client';
import { MessageReaction } from '../types';

export const reactionsAPI = {
  create: async (data: {
    message_id: string;
    reaction: string;
    application_id?: string;
  }): Promise<{ message: string; reaction: MessageReaction }> => {
    const response = await apiClient.post('/reactions', data);
    return response.data;
  },

  delete: async (data: {
    message_id: string;
    reaction: string;
    application_id?: string;
  }): Promise<{ message: string }> => {
    const response = await apiClient.delete('/reactions', { data });
    return response.data;
  },

  list: async (message_id: string, application_id?: string): Promise<{
    reactions: MessageReaction[];
    count: number;
  }> => {
    const params: any = { message_id };
    if (application_id) params.application_id = application_id;
    const response = await apiClient.get('/reactions', { params });
    return response.data;
  },
};


import { apiClient } from './client';
import { TypingIndicator } from '../types';

export const typingAPI = {
  create: async (data: {
    room_id: string;
    user_id: string;
    application_id?: string;
  }): Promise<{ typing: TypingIndicator }> => {
    const response = await apiClient.post('/typing', data);
    return response.data;
  },
};


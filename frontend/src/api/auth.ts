import { apiClient } from './client';
import { AuthResponse } from '../types';

export const authAPI = {
  register: async (data: {
    username: string;
    email: string;
    application_id?: string;
    external_id?: string;
    avatar_url?: string;
    metadata?: Record<string, any>;
  }): Promise<AuthResponse> => {
    const response = await apiClient.post('/auth/register', data);
    return response.data;
  },

  login: async (email: string): Promise<AuthResponse> => {
    const response = await apiClient.post('/auth/login', { email });
    return response.data;
  },

  refreshToken: async (refreshToken: string): Promise<{ access_token: string }> => {
    const response = await apiClient.post('/auth/refresh', { refresh_token: refreshToken });
    return response.data;
  },
};


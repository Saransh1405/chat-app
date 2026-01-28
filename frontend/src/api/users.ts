import { apiClient } from './client';
import { User } from '../types';

export const usersAPI = {
  create: async (data: {
    username: string;
    email: string;
    application_id?: string;
    external_id?: string;
    avatar_url?: string;
    metadata?: Record<string, any>;
  }): Promise<{ message: string; user: User }> => {
    const response = await apiClient.post('/users', data);
    return response.data;
  },

  get: async (id?: string, application_id?: string): Promise<{ user: User }> => {
    const params: any = {};
    if (id) params.id = id;
    if (application_id) params.application_id = application_id;
    const response = await apiClient.get('/users', { params });
    return response.data;
  },

  update: async (data: {
    id: string;
    application_id?: string;
    username?: string;
    email?: string;
    avatar_url?: string;
    metadata?: Record<string, any>;
  }): Promise<{ message: string; user: User }> => {
    const response = await apiClient.patch('/users', data);
    return response.data;
  },

  delete: async (id: string, application_id?: string): Promise<{ message: string }> => {
    const response = await apiClient.delete('/users', { data: { id, application_id } });
    return response.data;
  },
};


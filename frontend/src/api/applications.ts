import { apiClient } from './client';
import { Application } from '../types';

export const applicationsAPI = {
  create: async (data: { name: string }): Promise<{ message: string; application: Application & { secret_key: string } }> => {
    const response = await apiClient.post('/applications', data);
    return response.data;
  },

  get: async (id?: string): Promise<{ application: Application }> => {
    const params = id ? { id } : {};
    const response = await apiClient.get('/applications', { params });
    return response.data;
  },

  list: async (): Promise<{ applications: Application[]; count: number }> => {
    const response = await apiClient.get('/applications');
    return response.data;
  },

  update: async (data: { id: string; name?: string }): Promise<{ message: string; application: Application }> => {
    const response = await apiClient.patch('/applications', data);
    return response.data;
  },

  delete: async (id: string): Promise<{ message: string }> => {
    const response = await apiClient.delete('/applications', { data: { id } });
    return response.data;
  },
};


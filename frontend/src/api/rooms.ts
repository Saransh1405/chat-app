import { apiClient } from './client';
import { Room, RoomMember } from '../types';

export const roomsAPI = {
  create: async (data: {
    name: string;
    type?: 'group' | 'direct' | 'channel';
    description?: string;
    application_id?: string;
    created_by?: string;
    metadata?: Record<string, any>;
  }): Promise<{ message: string; room: Room }> => {
    const response = await apiClient.post('/rooms', data);
    return response.data;
  },

  get: async (id?: string, application_id?: string): Promise<{ room: Room }> => {
    const params: any = {};
    if (id) params.id = id;
    if (application_id) params.application_id = application_id;
    const response = await apiClient.get('/rooms', { params });
    return response.data;
  },

  list: async (application_id?: string): Promise<{ rooms: Room[]; count: number }> => {
    const params: any = {};
    if (application_id) params.application_id = application_id;
    const response = await apiClient.get('/rooms', { params });
    return response.data;
  },

  update: async (data: {
    id: string;
    application_id?: string;
    name?: string;
    description?: string;
    metadata?: Record<string, any>;
  }): Promise<{ message: string; room: Room }> => {
    const response = await apiClient.patch('/rooms', data);
    return response.data;
  },

  delete: async (id: string, application_id?: string): Promise<{ message: string }> => {
    const response = await apiClient.delete('/rooms', { data: { id, application_id } });
    return response.data;
  },

  addMember: async (data: {
    room_id: string;
    user_id: string;
    application_id?: string;
    role?: string;
  }): Promise<{ message: string; member: RoomMember }> => {
    const response = await apiClient.post('/rooms/members', data);
    return response.data;
  },

  removeMember: async (data: {
    room_id: string;
    user_id: string;
    application_id?: string;
  }): Promise<{ message: string }> => {
    const response = await apiClient.delete('/rooms/members', { data });
    return response.data;
  },

  listMembers: async (room_id: string, application_id?: string): Promise<{ members: RoomMember[]; count: number }> => {
    const params: any = { room_id };
    if (application_id) params.application_id = application_id;
    const response = await apiClient.get('/rooms/members', { params });
    return response.data;
  },
};


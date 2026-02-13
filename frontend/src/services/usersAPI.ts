import type { User } from "@/types/chat";
import { api } from "./api";

export const usersAPI = {
  async search(query: string, applicationId?: string): Promise<User[]> {
    const res = await api.get<{ users: User[] }>(`/users/all${applicationId ? `?application_id=${applicationId}` : ""}`);
    const users = res.users || [];
    if (!query) return users;
    return users.filter(
      (u) =>
        u.username.toLowerCase().includes(query.toLowerCase()) ||
        (u.email && u.email.toLowerCase().includes(query.toLowerCase()))
    );
  },

  async getById(userId: string): Promise<User> {
    const res = await api.get<{ user: User }>(`/users?id=${userId}`);
    return res.user;
  },

  async updateProfile(data: Partial<User>): Promise<User> {
    const res = await api.patch<{ user: User }>("/users", data);
    return res.user;
  },
};

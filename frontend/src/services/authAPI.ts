import type { AuthResponse, User } from "@/types/chat";
import { api } from "./api";

export const authAPI = {
  async login(email: string, _password?: string): Promise<AuthResponse> {
    return api.post<AuthResponse>("/auth/login", { email });
  },

  async signup(username: string, email: string, _password?: string): Promise<AuthResponse> {
    return api.post<AuthResponse>("/auth/register", {
      username,
      email,
    });
  },

  async me(): Promise<User> {
    const res = await api.get<{ user: User }>("/users");
    return res.user;
  },

  logout(): void {
    localStorage.removeItem("access_token");
    localStorage.removeItem("aura_user");
    localStorage.removeItem("refresh_token");
  },
};

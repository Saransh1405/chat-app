import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import type { User } from "@/types/chat";
import { authAPI } from "@/services/authAPI";
import { wsClient } from "@/services/wsClient";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password?: string) => Promise<void>;
  signup: (username: string, email: string, password?: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const savedToken = localStorage.getItem("access_token");
    const savedUser = localStorage.getItem("aura_user");

    if (savedToken && savedUser) {
      try {
        setToken(savedToken);
        setUser(JSON.parse(savedUser));
        wsClient.connect(savedToken);
      } catch {
        localStorage.removeItem("access_token");
        localStorage.removeItem("aura_user");
        localStorage.removeItem("refresh_token");
      }
    }
    setIsLoading(false);
  }, []);

  const login = useCallback(async (email: string, password?: string) => {
    const res = await authAPI.login(email, password);
    const accessToken = res.access_token;
    const refreshToken = res.refresh_token;
    const userData = res.user;

    if (accessToken) {
      localStorage.setItem("access_token", accessToken);
      setToken(accessToken);
      wsClient.connect(accessToken);
    }
    if (refreshToken) localStorage.setItem("refresh_token", refreshToken);
    if (userData) {
      localStorage.setItem("aura_user", JSON.stringify(userData));
      setUser(userData);
    }
  }, []);

  const signup = useCallback(async (username: string, email: string, password?: string) => {
    const res = await authAPI.signup(username, email, password);
    const accessToken = res.access_token;
    const refreshToken = res.refresh_token;
    const userData = res.user;

    if (accessToken) {
      localStorage.setItem("access_token", accessToken);
      setToken(accessToken);
      wsClient.connect(accessToken);
    }
    if (refreshToken) localStorage.setItem("refresh_token", refreshToken);
    if (userData) {
      localStorage.setItem("aura_user", JSON.stringify(userData));
      setUser(userData);
    }
  }, []);

  const logout = useCallback(() => {
    authAPI.logout();
    wsClient.disconnect();
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, token, isLoading, login, signup, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
};

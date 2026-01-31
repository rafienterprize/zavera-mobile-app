"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from "react";
import api from "@/lib/api";
import { triggerCartRefresh } from "./CartContext";

interface User {
  id: number;
  email: string;
  first_name: string;
  name?: string;
  phone?: string;
  birthdate?: string;
  is_verified: boolean;
  auth_provider: string;
  created_at: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<User>;
  loginWithGoogle: (idToken: string) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

interface RegisterData {
  first_name: string;
  email: string;
  password: string;
  birthdate: string;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    const storedToken = localStorage.getItem("auth_token");
    if (!storedToken) {
      setIsLoading(false);
      return;
    }

    setToken(storedToken);
    try {
      const response = await api.get("/auth/me", {
        headers: { Authorization: `Bearer ${storedToken}` },
      });
      setUser(response.data);
    } catch {
      localStorage.removeItem("auth_token");
      setToken(null);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (email: string, password: string) => {
    const response = await api.post("/auth/login", { email, password });
    const { access_token, user: userData } = response.data;
    localStorage.setItem("auth_token", access_token);
    setToken(access_token);
    setUser(userData);
    // Refresh cart after login to get user's persisted cart
    setTimeout(() => triggerCartRefresh(), 100);
    return userData; // Return user data for redirect logic
  };

  const loginWithGoogle = async (idToken: string) => {
    const response = await api.post("/auth/google", { id_token: idToken });
    const { access_token, user: userData } = response.data;
    localStorage.setItem("auth_token", access_token);
    setToken(access_token);
    setUser(userData);
    // Refresh cart after login to get user's persisted cart
    setTimeout(() => triggerCartRefresh(), 100);
  };

  const register = async (data: RegisterData) => {
    await api.post("/auth/register", data);
  };

  const logout = () => {
    localStorage.removeItem("auth_token");
    localStorage.removeItem("zavera_cart"); // Clear cart on logout
    setToken(null);
    setUser(null);
    // Trigger cart refresh to clear cart state
    setTimeout(() => triggerCartRefresh(), 100);
  };

  const refreshUser = async () => {
    const storedToken = localStorage.getItem("auth_token");
    if (!storedToken) return;

    try {
      const response = await api.get("/auth/me", {
        headers: { Authorization: `Bearer ${storedToken}` },
      });
      setUser(response.data);
    } catch {
      logout();
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isLoading,
        isAuthenticated: !!user,
        login,
        loginWithGoogle,
        register,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}

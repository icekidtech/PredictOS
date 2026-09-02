import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import api from "../services/api";

type User = { id: string; username: string; email: string; wallet_address?: string; avatar_url?: string; auth_provider?: string };

type AuthCtx = {
  user: User | null;
  token: string | null;
  login: (token: string, user: User) => void;
  logout: () => void;
  refreshUser: () => Promise<void>;
};

const Ctx = createContext<AuthCtx>(null!);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem("token"));
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem("user");
    return raw ? JSON.parse(raw) : null;
  });

  const login = (t: string, u: User) => {
    localStorage.setItem("token", t);
    localStorage.setItem("user", JSON.stringify(u));
    setToken(t);
    setUser(u);
  };
  const logout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    setToken(null);
    setUser(null);
  };

  const refreshUser = useCallback(async () => {
    if (!localStorage.getItem("token")) return;
    try {
      const res = await api.get("/auth/me");
      const u = res.data.user ?? res.data;
      localStorage.setItem("user", JSON.stringify(u));
      setUser(u);
    } catch {
      // token invalid — keep existing user
    }
  }, []);

  // Handle Google OAuth redirect ?token=...
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const t = params.get("token");
    if (t) {
      localStorage.setItem("token", t);
      setToken(t);
      window.history.replaceState({}, "", window.location.pathname);
      // Fetch fresh user after OAuth
      api.get("/auth/me").then((res) => {
        const u = res.data.user ?? res.data;
        localStorage.setItem("user", JSON.stringify(u));
        setUser(u);
      }).catch(() => {});
    } else if (token) {
      // On mount, refresh user to get latest wallet_address etc.
      refreshUser();
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return <Ctx.Provider value={{ user, token, login, logout, refreshUser }}>{children}</Ctx.Provider>;
}

export const useAuth = () => useContext(Ctx);

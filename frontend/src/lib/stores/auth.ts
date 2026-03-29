import { writable, derived } from "svelte/store";
import { api, setAccessToken, type ApiResponse } from "$lib/api/client";

export interface AuthUser {
  id: number;
  username: string;
  email?: string;
  iin?: string;
  role: string;
  language: string;
  created_at: string;
}

interface AuthState {
  user: AuthUser | null;
  loading: boolean;
}

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>({
    user: null,
    loading: true,
  });

  return {
    subscribe,

    async login(username: string, password: string): Promise<void> {
      const res = await api.post<{
        access_token: string;
        user: AuthUser;
      }>("/api/v1/auth/login", { username, password });

      setAccessToken(res.data.access_token);
      set({ user: res.data.user, loading: false });
    },

    async register(data: {
      username: string;
      password: string;
      email?: string;
      iin?: string;
      role: string;
      language?: string;
    }): Promise<void> {
      const res = await api.post<{
        access_token: string;
        user: AuthUser;
      }>("/api/v1/auth/register", data);

      setAccessToken(res.data.access_token);
      set({ user: res.data.user, loading: false });
    },

    async refresh(): Promise<boolean> {
      try {
        const res = await api.post<{
          access_token: string;
          user: AuthUser;
        }>("/api/v1/auth/refresh");

        setAccessToken(res.data.access_token);
        set({ user: res.data.user, loading: false });
        return true;
      } catch {
        setAccessToken(null);
        set({ user: null, loading: false });
        return false;
      }
    },

    async logout(): Promise<void> {
      try {
        await api.post("/api/v1/auth/logout");
      } catch {
        // Ignore errors on logout
      }
      setAccessToken(null);
      set({ user: null, loading: false });
    },

    reset() {
      setAccessToken(null);
      set({ user: null, loading: false });
    },
  };
}

export const auth = createAuthStore();

export const isAuthenticated = derived(auth, ($auth) => $auth.user !== null);
export const currentUser = derived(auth, ($auth) => $auth.user);
export const userRole = derived(auth, ($auth) => $auth.user?.role ?? null);
export const isLoading = derived(auth, ($auth) => $auth.loading);

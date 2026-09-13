import { create } from "zustand";
import { API_URL, getMe, type User } from "~/lib/api";

type AuthState = {
  user: User;
  isAuthenticated: boolean;
  isLoading: boolean;
  refreshSession: () => Promise<void>;
  login: (returnTo?: string) => void;
  register: (returnTo?: string) => void;
  logout: () => Promise<void>;
};

const emptyUser: User = { id: "" };

function authURL(path: string, returnTo = "/dashboard") {
  const query = new URLSearchParams({ return_to: returnTo });
  return `${API_URL}${path}?${query}`;
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: emptyUser,
  isAuthenticated: false,
  isLoading: true,
  refreshSession: async () => {
    set({ isLoading: true });
    try {
      const user = await getMe();
      set({ user: { ...user, firstName: user.username }, isAuthenticated: true, isLoading: false });
    } catch {
      set({ user: emptyUser, isAuthenticated: false, isLoading: false });
    }
  },
  login: (returnTo) => window.location.assign(authURL("/auth/login", returnTo)),
  register: (returnTo) => window.location.assign(authURL("/auth/register", returnTo)),
  logout: async () => {
    const response = await fetch(`${API_URL}/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
    if (!response.ok) throw new Error("logout failed");
    const result = (await response.json()) as { logout_url: string };
    set({ user: emptyUser, isAuthenticated: false, isLoading: false });
    window.location.assign(result.logout_url);
  },
}));

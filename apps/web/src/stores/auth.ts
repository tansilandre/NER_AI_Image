import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User, Organization } from '../types';
import { authApi } from '../lib/api';

interface AuthState {
  token: string | null;
  user: User | null;
  organization: Organization | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  
  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (data: {
    email: string;
    password: string;
    full_name: string;
    org_name: string;
  }) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  setAuth: (token: string, user: User, organization?: Organization) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      organization: null,
      isLoading: false,
      isAuthenticated: false,

      login: async (email, password) => {
        set({ isLoading: true });
        try {
          const response = await authApi.login(email, password);
          const { token, user, organization } = response.data;
          
          localStorage.setItem('token', token);
          set({
            token,
            user,
            organization,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      register: async (data) => {
        set({ isLoading: true });
        try {
          const response = await authApi.register(data);
          const { token, user, organization } = response.data;
          
          localStorage.setItem('token', token);
          set({
            token,
            user,
            organization,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: () => {
        localStorage.removeItem('token');
        set({
          token: null,
          user: null,
          organization: null,
          isAuthenticated: false,
        });
      },

      refreshToken: async () => {
        try {
          const response = await authApi.refresh();
          const { token } = response.data;
          
          localStorage.setItem('token', token);
          set({ token });
        } catch (error) {
          get().logout();
          throw error;
        }
      },

      setAuth: (token, user, organization) => {
        localStorage.setItem('token', token);
        set({
          token,
          user,
          organization,
          isAuthenticated: true,
        });
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        organization: state.organization,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

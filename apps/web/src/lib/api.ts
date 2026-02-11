import axios, { AxiosInstance } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:5005';

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authApi = {
  register: (data: {
    email: string;
    password: string;
    full_name: string;
    org_name: string;
  }) => api.post('/api/v1/auth/register', data),

  login: (email: string, password: string) =>
    api.post('/api/v1/auth/login', { email, password }),

  refresh: () => api.post('/api/v1/auth/refresh'),
};

// Generation API
export const generationApi = {
  create: (data: {
    base_prompt: string;
    provider_id: string;
    reference_images?: string[];
    product_images: string[];
    num_variations?: number;
  }) => api.post('/api/v1/generations', data),

  get: (id: string) => api.get(`/api/v1/generations/${id}`),

  list: (params?: { limit?: number; offset?: number }) =>
    api.get('/api/v1/generations', { params }),
};

// Gallery API
export const galleryApi = {
  list: (params?: { limit?: number; offset?: number }) =>
    api.get('/api/v1/gallery', { params }),
};

// Upload API
export const uploadApi = {
  upload: (file: File, folder: string) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('folder', folder);
    return api.post('/api/v1/uploads', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },
};

// Provider API
export const providerApi = {
  list: (category?: string) =>
    api.get('/api/v1/providers', { params: { category } }),
  
  listAll: () =>
    api.get('/api/v1/admin/providers'),
};

// Admin API
export const adminApi = {
  // Credits
  addCredits: (amount: number, note?: string) =>
    api.post('/api/v1/admin/credits', { amount, note }),

  listTransactions: (params?: { limit?: number; offset?: number }) =>
    api.get('/api/v1/admin/credits/history', { params }),

  // Members
  listMembers: () => api.get('/api/v1/admin/members'),

  inviteMember: (email: string) =>
    api.post('/api/v1/admin/members/invite', { email }),

  // Providers
  createProvider: (data: {
    name: string;
    slug: string;
    category: string;
    api_key?: string;
    model?: string;
    cost_per_use: number;
    is_active: boolean;
    config?: Record<string, any>;
  }) => api.post('/api/v1/admin/providers', data),

  updateProvider: (id: string, data: Partial<{
    name: string;
    api_key?: string;
    model?: string;
    cost_per_use: number;
    is_active: boolean;
    config?: Record<string, any>;
  }>) => api.patch(`/api/v1/admin/providers/${id}`, data),

  deleteProvider: (id: string) =>
    api.delete(`/api/v1/admin/providers/${id}`),

  testProvider: (slug: string) =>
    api.post(`/api/v1/admin/providers/${slug}/test`),
};

export default api;

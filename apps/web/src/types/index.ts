// User Types
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'member';
  organization_id: string;
  created_at: string;
}

export interface Organization {
  id: string;
  name: string;
  credits: number;
  created_at: string;
}

export interface Profile {
  id: string;
  email: string;
  full_name: string;
  avatar_url?: string;
  role: string;
  org_id: string;
  credits_spent: number;
  created_at: string;
  updated_at: string;
}

// Provider Types
export interface ProviderConfig {
  model?: string;
  default_aspect_ratio?: string;
  default_quality?: string;
  max_tokens?: number;
  temperature?: number;
  [key: string]: any;
}

export interface Provider {
  id: string;
  name: string;
  slug: string;
  category: 'image_generation' | 'llm' | 'vision';
  api_key: string;
  config: ProviderConfig;
  cost_per_use: number;
  is_active: boolean;
  priority: number;
  created_at: string;
  updated_at: string;
}

// Generation Types
export interface GenerationImage {
  id: string;
  generation_id: string;
  url: string;
  task_id?: string;
  status: 'pending' | 'generating' | 'completed' | 'failed';
  created_at: string;
  updated_at: string;
}

export interface Generation {
  id: string;
  user_id: string;
  org_id: string;
  provider_id: string;
  prompt: string;
  base_prompt: string;
  reference_images: string[];
  product_images: string[];
  num_variations: number;
  status: 'pending' | 'analyzing' | 'prompting' | 'generating' | 'completed' | 'failed';
  estimated_credits: number;
  actual_credits: number;
  total_images: number;
  completed_images: number;
  images?: GenerationImage[];
  error_message?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface CreateGenerationRequest {
  base_prompt: string;
  provider_id: string;
  reference_images?: string[];
  product_images: string[];
  num_variations?: number;
}

// Upload Types
export interface UploadResponse {
  url: string;
  key: string;
}

// Credit Types
export interface CreditTransaction {
  id: string;
  org_id: string;
  user_id: string;
  user_name: string;
  amount: number;
  balance: number;
  type: 'purchase' | 'usage' | 'refund';
  note?: string;
  generation_id?: string;
  created_at: string;
}

// Member Types
export interface Member {
  id: string;
  email: string;
  name: string;
  role: string;
  created_at: string;
  credits_spent: number;
}

// Admin Types
export interface CreateProviderRequest {
  name: string;
  slug: string;
  category: string;
  api_key?: string;
  model?: string;
  cost_per_use: number;
  is_active: boolean;
  config?: Record<string, any>;
}

export interface UpdateProviderRequest {
  name?: string;
  api_key?: string;
  model?: string;
  cost_per_use?: number;
  is_active?: boolean;
  config?: Record<string, any>;
}

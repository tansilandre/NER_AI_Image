import { create } from 'zustand';
import { Generation, Provider, GenerationImage } from '../types';
import { generationApi, providerApi, uploadApi } from '../lib/api';

interface UploadedImage {
  id: string;
  file: File;
  url?: string;
  status: 'uploading' | 'uploaded' | 'error';
}

interface GenerationState {
  // Images
  referenceImages: UploadedImage[];
  productImages: UploadedImage[];
  
  // Prompt
  prompt: string;
  
  // Settings
  imageCount: number;
  selectedProvider: Provider | null;
  availableProviders: Provider[];
  aspectRatio: string;
  quality: string;
  
  // Generation state
  isGenerating: boolean;
  currentGeneration: Generation | null;
  generatedImages: GenerationImage[];
  
  // Actions
  addReferenceImage: (file: File) => Promise<void>;
  addProductImage: (file: File) => Promise<void>;
  removeReferenceImage: (id: string) => void;
  removeProductImage: (id: string) => void;
  setPrompt: (prompt: string) => void;
  setImageCount: (count: number) => void;
  setProvider: (provider: Provider) => void;
  setAspectRatio: (ratio: string) => void;
  setQuality: (quality: string) => void;
  fetchProviders: () => Promise<void>;
  generate: () => Promise<void>;
  pollGeneration: (generationId: string) => void;
  getEstimatedCredits: () => number;
  reset: () => void;
}

export const useGenerationStore = create<GenerationState>((set, get) => ({
  referenceImages: [],
  productImages: [],
  prompt: '',
  imageCount: 4,
  selectedProvider: null,
  availableProviders: [],
  aspectRatio: '1:1',
  quality: 'standard',
  isGenerating: false,
  currentGeneration: null,
  generatedImages: [],

  addReferenceImage: async (file: File) => {
    const id = Math.random().toString(36).substring(7);
    const newImage: UploadedImage = { id, file, status: 'uploading' };
    
    set((state) => ({
      referenceImages: [...state.referenceImages, newImage],
    }));

    try {
      const response = await uploadApi.upload(file, 'references');
      const { url } = response.data;
      
      set((state) => ({
        referenceImages: state.referenceImages.map((img) =>
          img.id === id ? { ...img, url, status: 'uploaded' } : img
        ),
      }));
    } catch (error) {
      set((state) => ({
        referenceImages: state.referenceImages.map((img) =>
          img.id === id ? { ...img, status: 'error' } : img
        ),
      }));
    }
  },

  addProductImage: async (file: File) => {
    const id = Math.random().toString(36).substring(7);
    const newImage: UploadedImage = { id, file, status: 'uploading' };
    
    set((state) => ({
      productImages: [...state.productImages, newImage],
    }));

    try {
      const response = await uploadApi.upload(file, 'products');
      const { url } = response.data;
      
      set((state) => ({
        productImages: state.productImages.map((img) =>
          img.id === id ? { ...img, url, status: 'uploaded' } : img
        ),
      }));
    } catch (error) {
      set((state) => ({
        productImages: state.productImages.map((img) =>
          img.id === id ? { ...img, status: 'error' } : img
        ),
      }));
    }
  },

  removeReferenceImage: (id: string) => {
    set((state) => ({
      referenceImages: state.referenceImages.filter((img) => img.id !== id),
    }));
  },

  removeProductImage: (id: string) => {
    set((state) => ({
      productImages: state.productImages.filter((img) => img.id !== id),
    }));
  },

  setPrompt: (prompt: string) => set({ prompt }),
  
  setImageCount: (count: number) => set({ imageCount: count }),
  
  setProvider: (provider: Provider) => {
    const config = provider.config;
    set({
      selectedProvider: provider,
      aspectRatio: config.default_aspect_ratio || '1:1',
      quality: config.default_quality || 'standard',
    });
  },
  
  setAspectRatio: (ratio: string) => set({ aspectRatio: ratio }),
  
  setQuality: (quality: string) => set({ quality }),

  fetchProviders: async () => {
    try {
      const response = await providerApi.list('image_generation');
      const providers = response.data.providers || [];
      set({ 
        availableProviders: providers,
        selectedProvider: providers[0] || null,
      });
    } catch (error) {
      console.error('Failed to fetch providers:', error);
    }
  },

  generate: async () => {
    const { 
      prompt, 
      selectedProvider, 
      imageCount, 
      referenceImages, 
      productImages 
    } = get();

    if (!selectedProvider) {
      throw new Error('No provider selected');
    }

    if (!prompt.trim()) {
      throw new Error('Prompt is required');
    }

    if (productImages.length === 0) {
      throw new Error('At least one product image is required');
    }

    set({ isGenerating: true });

    try {
      const response = await generationApi.create({
        base_prompt: prompt,
        provider_id: selectedProvider.id,
        reference_images: referenceImages
          .filter((img) => img.status === 'uploaded')
          .map((img) => img.url!),
        product_images: productImages
          .filter((img) => img.status === 'uploaded')
          .map((img) => img.url!),
        num_variations: imageCount,
      });

      const generation: Generation = response.data;
      set({ 
        currentGeneration: generation,
        isGenerating: false,
      });

      // Start polling
      get().pollGeneration(generation.id);
    } catch (error) {
      set({ isGenerating: false });
      throw error;
    }
  },

  pollGeneration: (generationId: string) => {
    const poll = async () => {
      try {
        const response = await generationApi.get(generationId);
        const generation: Generation = response.data;
        
        set({ currentGeneration: generation });

        // Continue polling if not completed or failed
        if (
          generation.status !== 'completed' && 
          generation.status !== 'failed'
        ) {
          setTimeout(poll, 5000); // Poll every 5 seconds
        }
      } catch (error) {
        console.error('Failed to poll generation:', error);
      }
    };

    poll();
  },

  getEstimatedCredits: () => {
    const { imageCount, selectedProvider } = get();
    if (!selectedProvider) return 0;
    return imageCount * selectedProvider.cost_per_use;
  },

  reset: () => {
    set({
      referenceImages: [],
      productImages: [],
      prompt: '',
      imageCount: 4,
      currentGeneration: null,
      generatedImages: [],
    });
  },
}));

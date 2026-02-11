import { useEffect, useState } from 'react';
import { Sparkles, AlertCircle } from 'lucide-react';
import { Button, Card, TextArea, Badge } from '../components/ui';
import { ImageDropzone } from '../components';
import { useGenerationStore, useAuthStore } from '../stores';
import { toast } from 'sonner';

export function Generate() {
  const {
    referenceImages,
    productImages,
    prompt,
    imageCount,
    selectedProvider,
    availableProviders,
    isGenerating,
    currentGeneration,
    setPrompt,
    setImageCount,
    setProvider,
    addReferenceImage,
    addProductImage,
    removeReferenceImage,
    removeProductImage,
    fetchProviders,
    generate,
    getEstimatedCredits,
  } = useGenerationStore();

  const { organization } = useAuthStore();
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    fetchProviders();
  }, [fetchProviders]);

  const handleGenerate = async () => {
    if (!prompt.trim()) {
      toast.error('Please enter a prompt');
      return;
    }

    if (productImages.length === 0) {
      toast.error('Please upload at least one product image');
      return;
    }

    const estimatedCredits = getEstimatedCredits();
    if (organization && organization.credits < estimatedCredits) {
      toast.error('Insufficient credits');
      return;
    }

    setIsSubmitting(true);
    
    try {
      await generate();
      toast.success('Generation started!');
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to start generation');
    } finally {
      setIsSubmitting(false);
    }
  };

  const estimatedCredits = getEstimatedCredits();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading font-bold text-2xl">Generate Images</h1>
          <p className="text-gray-600">Create stunning product images with AI</p>
        </div>
        
        <div className="flex items-center gap-4">
          <div className="text-right">
            <p className="text-sm text-gray-500">Credits Available</p>
            <p className="font-heading font-bold text-lg">
              {organization?.credits || 0}
            </p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-6">
        {/* Left Column - Upload */}
        <div className="space-y-6">
          <Card padding="lg">
            <h3 className="font-heading font-bold mb-4">Reference Images</h3>
            <ImageDropzone
              label="Upload reference images (optional)"
              images={referenceImages}
              onAdd={addReferenceImage}
              onRemove={removeReferenceImage}
              maxFiles={5}
            />
            <p className="text-xs text-gray-500 mt-2">
              These help guide the AI on style and composition
            </p>
          </Card>

          <Card padding="lg">
            <h3 className="font-heading font-bold mb-4">Product Images *</h3>
            <ImageDropzone
              label="Upload product images"
              images={productImages}
              onAdd={addProductImage}
              onRemove={removeProductImage}
              maxFiles={3}
            />
            <p className="text-xs text-gray-500 mt-2">
              The product to be featured in generated images
            </p>
          </Card>
        </div>

        {/* Middle Column - Prompt */}
        <div className="col-span-2 space-y-6">
          <Card padding="lg">
            <h3 className="font-heading font-bold mb-4">Creative Brief</h3>
            
            <TextArea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Describe your vision for the generated images... e.g., 'A modern kitchen scene with the product on a marble countertop, natural lighting, warm tones'"
              className="h-40"
            />
            
            <div className="mt-4 flex items-center gap-2 text-sm text-gray-500">
              <AlertCircle className="w-4 h-4" />
              <span>
                {prompt.length} characters · Be specific about style, lighting, and setting
              </span>
            </div>
          </Card>

          {/* Settings */}
          <Card padding="lg">
            <h3 className="font-heading font-bold mb-4">Generation Settings</h3>
            
            <div className="grid grid-cols-3 gap-6">
              {/* Provider */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  AI Model
                </label>
                <select
                  value={selectedProvider?.id || ''}
                  onChange={(e) => {
                    const provider = availableProviders.find(
                      (p) => p.id === e.target.value
                    );
                    if (provider) setProvider(provider);
                  }}
                  className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none bg-white"
                >
                  {availableProviders.map((provider) => (
                    <option key={provider.id} value={provider.id}>
                      {provider.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Image Count */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Number of Variations
                </label>
                <select
                  value={imageCount}
                  onChange={(e) => setImageCount(Number(e.target.value))}
                  className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none bg-white"
                >
                  <option value={1}>1 image</option>
                  <option value={2}>2 images</option>
                  <option value={4}>4 images</option>
                  <option value={8}>8 images</option>
                </select>
              </div>

              {/* Aspect Ratio */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Aspect Ratio
                </label>
                <select
                  value={selectedProvider?.config.default_aspect_ratio || '1:1'}
                  className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none bg-white disabled:bg-gray-100"
                  disabled
                >
                  <option value="1:1">1:1 (Square)</option>
                  <option value="16:9">16:9 (Landscape)</option>
                  <option value="9:16">9:16 (Portrait)</option>
                  <option value="4:3">4:3</option>
                  <option value="3:4">3:4</option>
                </select>
              </div>
            </div>
          </Card>

          {/* Generate Button */}
          <div className="flex items-center justify-between p-6 bg-white shadow-md border-l-4 border-l-[var(--color-yellow)]">
            <div>
              <p className="text-sm text-gray-500">Estimated Cost</p>
              <p className="font-heading font-bold text-xl">
                {estimatedCredits} credits
              </p>
            </div>
            
            <Button
              onClick={handleGenerate}
              variant="primary"
              size="lg"
              isLoading={isSubmitting || isGenerating}
              disabled={productImages.length === 0 || !prompt.trim()}
            >
              <Sparkles className="w-5 h-5" />
              Generate Images
            </Button>
          </div>

          {/* Generation Status */}
          {currentGeneration && (
            <Card variant="accent">
              <h3 className="font-heading font-bold mb-2">Generation Status</h3>
              <div className="flex items-center gap-3">
                <Badge variant={
                  currentGeneration.status === 'completed' ? 'success' :
                  currentGeneration.status === 'failed' ? 'error' : 'warning'
                }>
                  {currentGeneration.status}
                </Badge>
                <span className="text-sm text-gray-600">
                  {currentGeneration.completed_images} / {currentGeneration.total_images} images completed
                </span>
              </div>
              <p className="text-sm text-gray-500 mt-2">
                Created: {new Date(currentGeneration.created_at).toLocaleString()}
              </p>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

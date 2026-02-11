import { useEffect, useState } from 'react';
import { Plus, Edit2, TestTube } from 'lucide-react';
import { Card, CardHeader, Button, Badge } from '../../components/ui';
import { adminApi, providerApi } from '../../lib/api';
import { Provider } from '../../types';
import { toast } from 'sonner';

export function Providers() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingProvider, setEditingProvider] = useState<Provider | null>(null);
  
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    category: 'image_generation',
    api_key: '',
    model: '',
    cost_per_use: 1,
    is_active: true,
    config: '',
  });

  const fetchProviders = async () => {
    try {
      setIsLoading(true);
      const response = await providerApi.listAll();
      setProviders(response.data.providers || []);
    } catch (error) {
      toast.error('Failed to fetch providers');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProviders();
  }, []);

  const handleSubmit = async () => {
    try {
      const data = {
        ...formData,
        config: formData.config ? JSON.parse(formData.config) : {},
      };

      if (editingProvider) {
        await adminApi.updateProvider(editingProvider.id, data);
        toast.success('Provider updated');
      } else {
        await adminApi.createProvider(data);
        toast.success('Provider created');
      }

      setShowAddModal(false);
      setEditingProvider(null);
      setFormData({
        name: '',
        slug: '',
        category: 'image_generation',
        api_key: '',
        model: '',
        cost_per_use: 1,
        is_active: true,
        config: '',
      });
      fetchProviders();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to save provider');
    }
  };

  const handleEdit = (provider: Provider) => {
    setEditingProvider(provider);
    setFormData({
      name: provider.name,
      slug: provider.slug,
      category: provider.category,
      api_key: '',
      model: provider.config.model || '',
      cost_per_use: provider.cost_per_use,
      is_active: provider.is_active,
      config: JSON.stringify(provider.config, null, 2),
    });
    setShowAddModal(true);
  };

  const handleTest = async (provider: Provider) => {
    try {
      await adminApi.testProvider(provider.slug);
      toast.success('Provider connection successful');
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Provider connection failed');
    }
  };

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'image_generation':
        return '🎨';
      case 'llm':
        return '🤖';
      case 'vision':
        return '👁️';
      default:
        return '⚙️';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading font-bold text-2xl">AI Providers</h1>
          <p className="text-gray-600">Manage AI service providers</p>
        </div>
        
        <Button 
          onClick={() => {
            setEditingProvider(null);
            setFormData({
              name: '',
              slug: '',
              category: 'image_generation',
              api_key: '',
              model: '',
              cost_per_use: 1,
              is_active: true,
              config: '',
            });
            setShowAddModal(true);
          }} 
          variant="primary"
        >
          <Plus className="w-4 h-4" />
          Add Provider
        </Button>
      </div>

      {/* Providers List */}
      <div className="grid grid-cols-2 gap-4">
        {providers.map((provider) => (
          <Card key={provider.id}>
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                <span className="text-2xl">{getCategoryIcon(provider.category)}</span>
                <div>
                  <h3 className="font-heading font-bold">{provider.name}</h3>
                  <p className="text-xs text-gray-500">{provider.slug}</p>
                </div>
              </div>
              
              <Badge variant={provider.is_active ? 'success' : 'default'}>
                {provider.is_active ? 'Active' : 'Inactive'}
              </Badge>
            </div>
            
            <div className="mt-4 space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-500">Category</span>
                <span className="font-medium capitalize">{provider.category}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-500">Model</span>
                <span className="font-medium">{provider.config.model || 'N/A'}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-500">Cost per use</span>
                <span className="font-medium">{provider.cost_per_use} credits</span>
              </div>
            </div>
            
            <div className="mt-4 flex gap-2">
              <Button
                variant="outline"
                size="sm"
                className="flex-1"
                onClick={() => handleEdit(provider)}
              >
                <Edit2 className="w-4 h-4" />
                Edit
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleTest(provider)}
              >
                <TestTube className="w-4 h-4" />
                Test
              </Button>
            </div>
          </Card>
        ))}
      </div>

      {/* Add/Edit Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
            <CardHeader 
              title={editingProvider ? 'Edit Provider' : 'Add Provider'} 
              subtitle="Configure AI service provider" 
            />
            
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Name
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                    placeholder="e.g., OpenAI GPT-4"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Slug
                  </label>
                  <input
                    type="text"
                    value={formData.slug}
                    onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                    className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                    placeholder="e.g., openai-gpt4"
                    disabled={!!editingProvider}
                  />
                </div>
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Category
                  </label>
                  <select
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none bg-white"
                  >
                    <option value="image_generation">Image Generation</option>
                    <option value="llm">LLM</option>
                    <option value="vision">Vision</option>
                  </select>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Model
                  </label>
                  <input
                    type="text"
                    value={formData.model}
                    onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                    className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                    placeholder="e.g., gpt-4"
                  />
                </div>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  API Key
                </label>
                <input
                  type="password"
                  value={formData.api_key}
                  onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
                  className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                  placeholder={editingProvider ? 'Leave blank to keep existing' : 'Enter API key'}
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Cost per use (credits)
                </label>
                <input
                  type="number"
                  value={formData.cost_per_use}
                  onChange={(e) => setFormData({ ...formData, cost_per_use: Number(e.target.value) })}
                  className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                  min={0}
                />
              </div>
              
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formData.is_active}
                  onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  className="w-4 h-4"
                />
                <label htmlFor="is_active" className="text-sm font-medium text-gray-700">
                  Active
                </label>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Config (JSON)
                </label>
                <textarea
                  value={formData.config}
                  onChange={(e) => setFormData({ ...formData, config: e.target.value })}
                  className="textarea w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none resize-none h-32"
                  placeholder="{}"
                />
              </div>
              
              <div className="flex gap-3 pt-4">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={() => setShowAddModal(false)}
                >
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  className="flex-1"
                  onClick={handleSubmit}
                  isLoading={isLoading}
                >
                  {editingProvider ? 'Update' : 'Create'}
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}

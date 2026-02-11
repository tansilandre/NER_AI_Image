import { useEffect, useState } from 'react';
import { Download, RefreshCw, Trash2 } from 'lucide-react';
import { Card, Badge, Button } from '../components/ui';
import { generationApi } from '../lib/api';
import { Generation, GenerationImage } from '../types';
import { toast } from 'sonner';

export function Gallery() {
  const [generations, setGenerations] = useState<Generation[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchGenerations = async () => {
    try {
      setIsLoading(true);
      const response = await generationApi.list();
      setGenerations(response.data.generations || []);
    } catch (error) {
      toast.error('Failed to fetch generations');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchGenerations();
  }, []);

  const getStatusVariant = (status: string) => {
    switch (status) {
      case 'completed':
        return 'success';
      case 'failed':
        return 'error';
      case 'pending':
        return 'warning';
      default:
        return 'default';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading font-bold text-2xl">Gallery</h1>
          <p className="text-gray-600">View your generated images</p>
        </div>
        
        <Button
          variant="outline"
          onClick={fetchGenerations}
          isLoading={isLoading}
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </Button>
      </div>

      {/* Generations List */}
      {generations.length === 0 ? (
        <Card className="text-center py-16">
          <div className="max-w-md mx-auto">
            <div className="w-16 h-16 bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <RefreshCw className="w-8 h-8 text-gray-400" />
            </div>
            <h3 className="font-heading font-bold text-lg mb-2">
              No generations yet
            </h3>
            <p className="text-gray-500 mb-4">
              Start generating images and they will appear here.
            </p>
            <Button onClick={() => window.location.href = '/generate'}>
              Start Generating
            </Button>
          </div>
        </Card>
      ) : (
        <div className="space-y-4">
          {generations.map((generation) => (
            <Card key={generation.id} className="overflow-hidden">
              <div className="p-4 border-b border-gray-100 flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <Badge variant={getStatusVariant(generation.status)}>
                    {generation.status}
                  </Badge>
                  <span className="text-sm text-gray-500">
                    {new Date(generation.created_at).toLocaleString()}
                  </span>
                  <span className="text-sm text-gray-500">
                    {generation.completed_images} / {generation.total_images} images
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <button className="p-2 text-gray-400 hover:text-red-500 transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
              
              <div className="p-4">
                <p className="text-sm text-gray-700 mb-4 line-clamp-2">
                  {generation.prompt}
                </p>
                
                {generation.images && generation.images.length > 0 && (
                  <div className="grid grid-cols-4 gap-4">
                    {generation.images.map((image: GenerationImage) => (
                      <div key={image.id} className="group relative aspect-square bg-gray-100">
                        <img
                          src={image.url}
                          alt="Generated"
                          className="w-full h-full object-cover"
                          loading="lazy"
                        />
                        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                          <a
                            href={image.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="w-10 h-10 bg-white flex items-center justify-center shadow-lg hover:bg-gray-100"
                          >
                            <Download className="w-5 h-5" />
                          </a>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

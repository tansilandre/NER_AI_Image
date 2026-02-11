import { useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { Upload, X } from 'lucide-react';

interface UploadedFile {
  id: string;
  file?: File;
  url?: string;
  status: 'uploading' | 'uploaded' | 'error';
}

interface ImageDropzoneProps {
  label: string;
  images: UploadedFile[];
  onAdd: (file: File) => Promise<void>;
  onRemove: (id: string) => void;
  accept?: string;
  maxFiles?: number;
}

export function ImageDropzone({
  label,
  images,
  onAdd,
  onRemove,
  accept = 'image/*',
  maxFiles = 5,
}: ImageDropzoneProps) {
  const onDrop = useCallback(
    async (acceptedFiles: File[]) => {
      for (const file of acceptedFiles) {
        if (images.length >= maxFiles) break;
        await onAdd(file);
      }
    },
    [images.length, maxFiles, onAdd]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { 'image/*': ['.png', '.jpg', '.jpeg', '.webp'] },
    maxFiles: maxFiles - images.length,
    disabled: images.length >= maxFiles,
  });

  return (
    <div className="space-y-3">
      <p className="text-sm font-medium text-gray-700">{label}</p>
      
      {/* Dropzone */}
      <div
        {...getRootProps()}
        className={`border-2 border-dashed border-gray-300 p-6 text-center cursor-pointer transition-colors ${
          isDragActive ? 'border-[var(--color-yellow)] bg-[var(--color-yellow)]/5' : 'hover:border-gray-400'
        } ${images.length >= maxFiles ? 'opacity-50 cursor-not-allowed' : ''}`}
      >
        <input {...getInputProps()} accept={accept} />
        <Upload className="w-8 h-8 mx-auto mb-2 text-gray-400" />
        <p className="text-sm text-gray-600">
          {isDragActive
            ? 'Drop the images here...'
            : 'Drag & drop images here, or click to select'}
        </p>
        <p className="text-xs text-gray-400 mt-1">
          PNG, JPG, JPEG, WEBP up to 10MB
        </p>
      </div>

      {/* Image Previews */}
      {images.length > 0 && (
        <div className="grid grid-cols-4 gap-3">
          {images.map((image) => (
            <div
              key={image.id}
              className="relative aspect-square bg-gray-100 border border-gray-200"
            >
              {image.file && (
                <img
                  src={URL.createObjectURL(image.file)}
                  alt="Preview"
                  className="w-full h-full object-cover"
                />
              )}
              
              {/* Status Overlay */}
              {image.status === 'uploading' && (
                <div className="absolute inset-0 bg-black/50 flex items-center justify-center">
                  <span className="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin" />
                </div>
              )}
              
              {image.status === 'error' && (
                <div className="absolute inset-0 bg-red-500/50 flex items-center justify-center">
                  <span className="text-white text-xs">Error</span>
                </div>
              )}
              
              {/* Remove Button */}
              <button
                onClick={() => onRemove(image.id)}
                className="absolute top-1 right-1 w-6 h-6 bg-white shadow flex items-center justify-center hover:bg-gray-100 transition-colors"
              >
                <X className="w-3 h-3" />
              </button>
            </div>
          ))}
        </div>
      )}
      
      <p className="text-xs text-gray-500">
        {images.length} / {maxFiles} images
      </p>
    </div>
  );
}

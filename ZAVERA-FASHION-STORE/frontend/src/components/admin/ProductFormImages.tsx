'use client';

import { useState } from 'react';
import { Upload, X, Star, MoveUp, MoveDown } from 'lucide-react';
import { useDialog } from '@/hooks/useDialog';

interface Props {
  formData: any;
  setFormData: (data: any) => void;
}

export default function ProductFormImages({ formData, setFormData }: Props) {
  const dialog = useDialog();
  const [uploading, setUploading] = useState(false);
  const [dragActive, setDragActive] = useState(false);

  const handleImageUpload = async (files: FileList) => {
    if (files.length === 0) return;

    setUploading(true);
    try {
      const token = localStorage.getItem('auth_token');
      const uploadPromises = Array.from(files).map(async (file) => {
        const formDataUpload = new FormData();
        formDataUpload.append('image', file);

        const response = await fetch('http://localhost:8080/api/admin/products/upload-image', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
          body: formDataUpload,
        });

        if (!response.ok) throw new Error('Upload failed');
        const data = await response.json();
        return data.image_url;
      });

      const urls = await Promise.all(uploadPromises);
      setFormData({
        ...formData,
        images: [...formData.images, ...urls],
      });
    } catch (error) {
      console.error('Failed to upload images:', error);
      await dialog.alert({
        title: 'Error',
        message: 'Gagal mengupload gambar',
      });
    } finally {
      setUploading(false);
    }
  };

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleImageUpload(e.dataTransfer.files);
    }
  };

  const removeImage = (index: number) => {
    setFormData({
      ...formData,
      images: formData.images.filter((_: any, i: number) => i !== index),
    });
  };

  const moveImage = (index: number, direction: 'up' | 'down') => {
    const newImages = [...formData.images];
    const newIndex = direction === 'up' ? index - 1 : index + 1;
    if (newIndex < 0 || newIndex >= newImages.length) return;

    [newImages[index], newImages[newIndex]] = [newImages[newIndex], newImages[index]];
    setFormData({ ...formData, images: newImages });
  };

  const setPrimaryImage = (index: number) => {
    const newImages = [...formData.images];
    const [primary] = newImages.splice(index, 1);
    newImages.unshift(primary);
    setFormData({ ...formData, images: newImages });
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Product Images</h2>
        <p className="text-gray-600 mt-1">Upload multiple images. First image will be the primary image.</p>
      </div>

      {/* Upload Area */}
      <div
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
        className={`relative border-2 border-dashed rounded-xl p-12 text-center transition-all ${
          dragActive ? 'border-emerald-500 bg-emerald-50' : 'border-gray-300 hover:border-gray-400'
        }`}
      >
        <input
          type="file"
          accept="image/jpeg,image/jpg,image/png,image/webp"
          multiple
          onChange={(e) => e.target.files && handleImageUpload(e.target.files)}
          disabled={uploading}
          className="hidden"
          id="file-upload"
        />
        <label htmlFor="file-upload" className="cursor-pointer">
          <div className="flex flex-col items-center gap-4">
            <div className="w-20 h-20 rounded-full bg-emerald-100 flex items-center justify-center">
              <Upload className="text-emerald-600" size={32} />
            </div>
            <div>
              <p className="text-lg font-medium mb-1">
                {uploading ? 'Uploading...' : 'Upload Product Images'}
              </p>
              <p className="text-gray-500 text-sm">Drag and drop or click to browse</p>
              <p className="text-gray-400 text-xs mt-2">
                JPG, PNG, WEBP (Max 5MB per image) • Recommended: 1000x1000px
              </p>
            </div>
          </div>
        </label>
      </div>

      {/* Image Gallery */}
      {formData.images.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">
              Uploaded Images ({formData.images.length})
            </h3>
            <p className="text-sm text-gray-500">Drag to reorder or use arrows</p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {formData.images.map((url: string, index: number) => (
              <div key={index} className="relative group">
                <div className="aspect-square rounded-lg overflow-hidden border-2 border-gray-200 group-hover:border-emerald-500 transition-colors">
                  <img src={url} alt={`Product ${index + 1}`} className="w-full h-full object-cover" />
                </div>

                {/* Primary Badge */}
                {index === 0 && (
                  <div className="absolute top-2 left-2 px-2 py-1 bg-emerald-500 text-white text-xs font-bold rounded flex items-center gap-1">
                    <Star size={12} fill="white" />
                    PRIMARY
                  </div>
                )}

                {/* Actions */}
                <div className="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  {index > 0 && (
                    <button
                      onClick={() => moveImage(index, 'up')}
                      className="p-1.5 bg-white rounded shadow hover:bg-gray-100"
                      title="Move up"
                    >
                      <MoveUp size={16} />
                    </button>
                  )}
                  {index < formData.images.length - 1 && (
                    <button
                      onClick={() => moveImage(index, 'down')}
                      className="p-1.5 bg-white rounded shadow hover:bg-gray-100"
                      title="Move down"
                    >
                      <MoveDown size={16} />
                    </button>
                  )}
                  <button
                    onClick={() => removeImage(index)}
                    className="p-1.5 bg-red-500 text-white rounded shadow hover:bg-red-600"
                    title="Remove"
                  >
                    <X size={16} />
                  </button>
                </div>

                {/* Set as Primary */}
                {index !== 0 && (
                  <button
                    onClick={() => setPrimaryImage(index)}
                    className="absolute bottom-2 left-2 right-2 px-2 py-1 bg-white/90 backdrop-blur text-xs font-medium rounded opacity-0 group-hover:opacity-100 transition-opacity hover:bg-white"
                  >
                    Set as Primary
                  </button>
                )}

                {/* Image Number */}
                <div className="absolute bottom-2 right-2 px-2 py-1 bg-black/70 text-white text-xs rounded">
                  #{index + 1}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Requirements */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h4 className="font-semibold text-blue-900 mb-2">Image Requirements:</h4>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>• Minimum 3 images required</li>
          <li>• First image will be used as the main product image</li>
          <li>• Use high-quality images (min 1000x1000px)</li>
          <li>• Show product from different angles</li>
          <li>• Include close-up shots of details</li>
          <li>• Use white or neutral background for best results</li>
        </ul>
      </div>

      {formData.images.length < 3 && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 flex items-start gap-3">
          <div className="text-amber-600 mt-0.5">⚠️</div>
          <div>
            <p className="font-medium text-amber-900">More images needed</p>
            <p className="text-sm text-amber-800">
              Upload at least {3 - formData.images.length} more image(s) to continue
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

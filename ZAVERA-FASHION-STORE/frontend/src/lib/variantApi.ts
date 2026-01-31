import {
  ProductVariant,
  ProductWithVariants,
  CreateVariantRequest,
  BulkGenerateVariantsRequest,
  VariantImage,
  LowStockVariant,
  VariantStockSummary,
  VariantAttribute,
  AvailableOptions,
} from '@/types/variant';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export const variantApi = {
  // Public endpoints
  async getProductVariants(productId: number): Promise<ProductVariant[]> {
    const res = await fetch(`${API_URL}/products/${productId}/variants`);
    if (!res.ok) throw new Error('Failed to fetch variants');
    const data = await res.json();
    // Handle both array and wrapped response
    return Array.isArray(data) ? data : (data.value || data.variants || []);
  },

  async getProductWithVariants(productId: number): Promise<ProductWithVariants> {
    const res = await fetch(`${API_URL}/products/${productId}/with-variants`);
    if (!res.ok) throw new Error('Failed to fetch product with variants');
    return res.json();
  },

  async getAvailableOptions(productId: number): Promise<AvailableOptions> {
    const res = await fetch(`${API_URL}/products/${productId}/options`);
    if (!res.ok) throw new Error('Failed to fetch options');
    return res.json();
  },

  async getVariant(id: number): Promise<ProductVariant> {
    const res = await fetch(`${API_URL}/variants/${id}`);
    if (!res.ok) throw new Error('Failed to fetch variant');
    return res.json();
  },

  async getVariantBySKU(sku: string): Promise<ProductVariant> {
    const res = await fetch(`${API_URL}/variants/sku/${sku}`);
    if (!res.ok) throw new Error('Failed to fetch variant');
    return res.json();
  },

  async getVariantImages(variantId: number): Promise<VariantImage[]> {
    const res = await fetch(`${API_URL}/variants/${variantId}/images`);
    if (!res.ok) throw new Error('Failed to fetch images');
    return res.json();
  },

  async checkAvailability(variantId: number, quantity: number): Promise<{
    available: boolean;
    available_stock: number;
    requested_stock: number;
  }> {
    const res = await fetch(`${API_URL}/variants/check-availability`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ variant_id: variantId, quantity }),
    });
    if (!res.ok) throw new Error('Failed to check availability');
    return res.json();
  },

  async getVariantAttributes(): Promise<VariantAttribute[]> {
    const res = await fetch(`${API_URL}/variants/attributes`);
    if (!res.ok) throw new Error('Failed to fetch attributes');
    return res.json();
  },

  async findVariant(productId: number, size?: string, color?: string): Promise<ProductVariant> {
    const res = await fetch(`${API_URL}/products/variants/find`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ product_id: productId, size, color }),
    });
    if (!res.ok) throw new Error('Variant not found');
    return res.json();
  },

  // Admin endpoints
  async createVariant(token: string, data: CreateVariantRequest): Promise<ProductVariant> {
    const res = await fetch(`${API_URL}/admin/variants`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error('Failed to create variant');
    return res.json();
  },

  async updateVariant(token: string, id: number, data: Partial<CreateVariantRequest>): Promise<ProductVariant> {
    console.log('🌐 variantApi.updateVariant called');
    console.log('📝 ID:', id);
    console.log('📝 Data:', data);
    
    const res = await fetch(`${API_URL}/admin/variants/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });
    
    console.log('📡 Response status:', res.status);
    console.log('📡 Response OK:', res.ok);
    
    if (!res.ok) {
      const errorText = await res.text();
      console.error('❌ Response error:', errorText);
      throw new Error('Failed to update variant');
    }
    
    const result = await res.json();
    console.log('✅ Response data:', result);
    return result;
  },

  async deleteVariant(token: string, id: number): Promise<void> {
    const res = await fetch(`${API_URL}/admin/variants/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error('Failed to delete variant');
  },

  async bulkGenerateVariants(token: string, data: BulkGenerateVariantsRequest): Promise<{
    message: string;
    count: number;
    variants: ProductVariant[];
  }> {
    const res = await fetch(`${API_URL}/admin/variants/bulk-generate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error('Failed to generate variants');
    return res.json();
  },

  async addVariantImage(token: string, variantId: number, imageUrl: string, isPrimary = false): Promise<VariantImage> {
    const res = await fetch(`${API_URL}/admin/variants/images`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        variant_id: variantId,
        image_url: imageUrl,
        is_primary: isPrimary,
        position: 0,
      }),
    });
    if (!res.ok) throw new Error('Failed to add image');
    return res.json();
  },

  async deleteVariantImage(token: string, imageId: number): Promise<void> {
    const res = await fetch(`${API_URL}/admin/variants/images/${imageId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error('Failed to delete image');
  },

  async setPrimaryImage(token: string, variantId: number, imageId: number): Promise<void> {
    const res = await fetch(`${API_URL}/admin/variants/${variantId}/images/primary`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ image_id: imageId }),
    });
    if (!res.ok) throw new Error('Failed to set primary image');
  },

  async reorderImages(token: string, variantId: number, imageIds: number[]): Promise<void> {
    const res = await fetch(`${API_URL}/admin/variants/${variantId}/images/reorder`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ image_ids: imageIds }),
    });
    if (!res.ok) throw new Error('Failed to reorder images');
  },

  async updateStock(token: string, variantId: number, quantity: number): Promise<void> {
    const res = await fetch(`${API_URL}/admin/variants/stock/${variantId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ quantity }),
    });
    if (!res.ok) throw new Error('Failed to update stock');
  },

  async adjustStock(token: string, variantId: number, delta: number): Promise<void> {
    const res = await fetch(`${API_URL}/admin/variants/stock/${variantId}/adjust`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ delta }),
    });
    if (!res.ok) throw new Error('Failed to adjust stock');
  },

  async getLowStockVariants(token: string): Promise<LowStockVariant[]> {
    const res = await fetch(`${API_URL}/admin/variants/low-stock`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error('Failed to fetch low stock variants');
    return res.json();
  },

  async getStockSummary(token: string, productId: number): Promise<VariantStockSummary[]> {
    const res = await fetch(`${API_URL}/admin/variants/stock-summary/${productId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error('Failed to fetch stock summary');
    return res.json();
  },
};

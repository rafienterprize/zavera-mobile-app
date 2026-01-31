export type ProductCategory = 'wanita' | 'pria' | 'anak' | 'sports' | 'luxury' | 'beauty';

export interface Product {
  id: number;
  name: string;
  price: number;
  description: string;
  image_url?: string; // Primary image (legacy/fallback)
  images?: string[]; // Array of image URLs
  stock: number;
  weight?: number;
  category: ProductCategory;
  subcategory?: string;
  brand?: string; // Product brand (e.g., Nike, Adidas)
  material?: string; // Product material (e.g., Cotton, Polyester)
  available_sizes?: string[]; // Available sizes from variants
}

export interface CartItem extends Product {
  quantity: number;
  selectedSize?: string;
  cartItemId?: number; // Backend cart item ID for sync
}

export interface CheckoutRequest {
  customer_name: string;
  email: string;
  phone: string;
  items: {
    product_id: number;
    quantity: number;
  }[];
}

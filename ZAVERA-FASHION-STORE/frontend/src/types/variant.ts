export interface VariantAttributes {
  [key: string]: any;
}

export interface ProductVariant {
  id: number;
  product_id: number;
  sku: string;
  variant_name: string;
  size?: string;
  color?: string;
  color_hex?: string;
  material?: string;
  pattern?: string;
  fit?: string;
  sleeve?: string;
  custom_attributes?: VariantAttributes;
  price?: number;
  compare_at_price?: number;
  cost_per_item?: number;
  stock_quantity: number;
  reserved_stock: number;
  low_stock_threshold: number;
  is_active: boolean;
  is_default: boolean;
  weight_grams?: number;
  weight?: number; // Alias for weight_grams
  length_cm?: number; // From backend
  length?: number; // Alias for length_cm
  width_cm?: number; // From backend
  width?: number; // Alias for width_cm
  height_cm?: number; // From backend
  height?: number; // Alias for height_cm
  barcode?: string;
  position: number;
  created_at: string;
  updated_at: string;
  images?: VariantImage[];
  available_stock?: number;
}

export interface VariantImage {
  id: number;
  variant_id: number;
  image_url: string;
  alt_text?: string;
  position: number;
  is_primary: boolean;
  width?: number;
  height?: number;
  format?: string;
  created_at: string;
}

export interface VariantAttribute {
  id: number;
  name: string;
  display_name: string;
  type: 'size' | 'color' | 'text' | 'select';
  options?: string[];
  sort_order: number;
  is_active: boolean;
  created_at: string;
}

export interface ProductWithVariants {
  id: number;
  name: string;
  description: string;
  price: number;
  category: string;
  image_url: string;
  stock: number;
  variants: ProductVariant[];
  price_range?: {
    min_price: number;
    max_price: number;
  };
}

export interface LowStockVariant {
  id: number;
  product_id: number;
  product_name: string;
  sku: string;
  variant_name: string;
  size?: string;
  color?: string;
  stock_quantity: number;
  low_stock_threshold: number;
  available_stock: number;
}

export interface VariantStockSummary {
  variant_id: number;
  product_id: number;
  product_name: string;
  sku: string;
  variant_name: string;
  stock_quantity: number;
  reserved_quantity: number;
  available_quantity: number;
}

export interface CreateVariantRequest {
  product_id: number;
  sku?: string;
  variant_name?: string;
  size?: string;
  color?: string;
  color_hex?: string;
  material?: string;
  pattern?: string;
  fit?: string;
  sleeve?: string;
  custom_attributes?: VariantAttributes;
  price?: number;
  compare_at_price?: number;
  cost_per_item?: number;
  stock_quantity: number;
  low_stock_threshold?: number;
  is_active: boolean;
  is_default?: boolean;
  weight_grams?: number;
  weight?: number; // Weight in grams
  length?: number; // Length in cm
  width?: number; // Width in cm
  height?: number; // Height in cm
  barcode?: string;
  position?: number;
}

export interface BulkGenerateVariantsRequest {
  product_id: number;
  sizes: string[];
  colors: string[];
  base_price?: number;
  stock_per_variant: number;
  weight?: number; // Default weight in grams
  length?: number; // Default length in cm
  width?: number; // Default width in cm
  height?: number; // Default height in cm
}

export interface AvailableOptions {
  size?: string[];
  color?: string[];
  material?: string[];
  pattern?: string[];
  fit?: string[];
  sleeve?: string[];
}

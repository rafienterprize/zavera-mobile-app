// Shipping Types

export interface Province {
  province_id: string;
  province_name: string;
}

export interface City {
  city_id: string;
  city_name: string;
  province_id: string;
  province_name: string;
  type: string;
  postal_code: string;
}

export interface Subdistrict {
  id: number;
  city_id: string;
  name: string;
  postal_codes: string[];
}

// District from RajaOngkir API - required for shipping calculation
export interface District {
  district_id: string;
  district_name: string;
  city_id?: string;
  postal_codes?: string[];
}

// Kelurahan from RajaOngkir API - provides postal codes
export interface Kelurahan {
  subdistrict_id: string;
  subdistrict_name: string;
  district_id: string;
  postal_code: string;
}

// Biteship Area - new area-based location system
export interface BiteshipArea {
  area_id: string;
  name: string;           // Full path: "Kelurahan, Kecamatan, Kota, Provinsi"
  postal_code: string;
  province?: string;
  city?: string;
  district?: string;
  subdistrict?: string;
}

// Shipping Category type - matches Tokopedia/Shopee grouping
export type ShippingCategory = 'Express' | 'Regular' | 'Economy' | 'SameDay';

export interface ShippingRate {
  provider_code: string;
  provider_name: string;
  provider_logo: string;
  service_code: string;
  service_name: string;
  description: string;
  cost: number;
  etd: string;                    // Raw ETD: "1-2"
  eta_date: string;               // Formatted: "Tiba 12 - 13 Jan"
  shipping_category: ShippingCategory;
  is_absurd_price?: boolean;      // Price > 5x REG price
}

export interface ShippingAddress {
  recipient_name: string;
  phone: string;
  province_id?: string;
  province_name?: string;
  city_id?: string;       // Optional - legacy field
  city_name: string;
  district_id?: string;   // Optional - legacy field
  district?: string;
  subdistrict?: string;
  postal_code?: string;
  full_address: string;
  area_id?: string;       // Biteship area_id for new system
}

// Enhanced cart shipping preview response (Tokopedia/Shopee style)
export interface CartShippingPreview {
  cart_subtotal: number;
  total_weight: number;           // in grams
  total_weight_kg: string;        // "1.2 kg"
  origin_city: string;            // "Semarang"
  destination_city?: string;
  grouped_rates: Record<string, ShippingRate[]>;  // Grouped by category
  rates: ShippingRate[];          // Flat list, sorted
  regular_min_price: number;      // For absurd price reference
}

export interface CheckoutWithShippingRequest {
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  notes?: string;
  address_id?: number;
  shipping_address?: ShippingAddress;
  provider_code: string;
  service_code: string;
}

export interface CheckoutWithShippingResponse {
  order_id: number;
  order_code: string;
  subtotal: number;
  shipping_cost: number;
  total_amount: number;
  status: string;
  shipping_locked: boolean;
  provider: string;
  service: string;
  etd: string;
  shipping_address: {
    recipient_name: string;
    phone: string;
    full_address: string;
    city_name: string;
    province_name: string;
    postal_code: string;
  };
}

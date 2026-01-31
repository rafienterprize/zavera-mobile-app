# Design Document

## Overview

Redesign UI/UX ZAVERA Fashion Store dengan konsep premium dan elegan. Fokus utama adalah pemisahan collection berdasarkan kategori fashion dengan navigasi yang intuitif, sambil mempertahankan integrasi backend dan payment gateway Midtrans yang sudah ada.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Next.js)                      │
├─────────────────────────────────────────────────────────────┤
│  Pages:                                                      │
│  ├── / (Homepage - Hero, Categories, Featured)              │
│  ├── /wanita (Women's Collection)                           │
│  ├── /pria (Men's Collection)                               │
│  ├── /anak (Kids Collection)                                │
│  ├── /sports (Sports Collection)                            │
│  ├── /luxury (Luxury Collection)                            │
│  ├── /beauty (Beauty Collection)                            │
│  ├── /product/[id] (Product Detail)                         │
│  ├── /cart (Shopping Cart)                                  │
│  └── /checkout (Checkout - existing)                        │
├─────────────────────────────────────────────────────────────┤
│  Components:                                                 │
│  ├── Header (Premium navigation with mega-menu)             │
│  ├── CategoryPage (Reusable collection template)            │
│  ├── ProductCard (Premium product display)                  │
│  ├── HeroBanner (Homepage hero section)                     │
│  └── CategoryGrid (Category navigation cards)               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Go/Gin)                          │
├─────────────────────────────────────────────────────────────┤
│  Endpoints:                                                  │
│  ├── GET /products?category={category}                      │
│  ├── GET /products/:id                                      │
│  ├── POST /cart (existing)                                  │
│  ├── POST /checkout (existing)                              │
│  └── POST /webhook/midtrans (existing)                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Database (PostgreSQL)                      │
├─────────────────────────────────────────────────────────────┤
│  products table + category column                            │
│  Categories: wanita, pria, anak, sports, luxury, beauty     │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Premium Header Component
```typescript
interface HeaderProps {
  transparent?: boolean;
}

// Categories for navigation
const categories = [
  { name: 'WANITA', href: '/wanita', subcategories: ['Dress', 'Tops', 'Bottoms', 'Outerwear'] },
  { name: 'PRIA', href: '/pria', subcategories: ['Shirts', 'Pants', 'Jackets', 'Suits'] },
  { name: 'SPORTS', href: '/sports', subcategories: ['Activewear', 'Footwear', 'Accessories'] },
  { name: 'ANAK', href: '/anak', subcategories: ['Boys', 'Girls', 'Baby'] },
  { name: 'LUXURY', href: '/luxury', subcategories: ['Designer', 'Premium', 'Limited Edition'] },
  { name: 'BEAUTY', href: '/beauty', subcategories: ['Skincare', 'Makeup', 'Fragrance'] },
];
```

### 2. Category Page Component
```typescript
interface CategoryPageProps {
  category: string;
  title: string;
  description: string;
  bannerImage: string;
}
```

### 3. Product Card Component (Enhanced)
```typescript
interface ProductCardProps {
  product: Product;
  variant?: 'default' | 'luxury';
}
```

## Data Models

### Updated Product Type
```typescript
export interface Product {
  id: number;
  name: string;
  price: number;
  description: string;
  image_url: string;
  stock: number;
  category: 'wanita' | 'pria' | 'anak' | 'sports' | 'luxury' | 'beauty';
  subcategory?: string;
}
```

### Database Migration
```sql
ALTER TABLE products ADD COLUMN category VARCHAR(50) DEFAULT 'wanita';
ALTER TABLE products ADD COLUMN subcategory VARCHAR(100);
CREATE INDEX idx_products_category ON products(category);
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Category Filter Consistency
*For any* category page request, all returned products SHALL have a category field matching the requested category.
**Validates: Requirements 1.2, 5.2**

### Property 2: Cart Persistence Across Pages
*For any* product added to cart from any collection page, the cart state SHALL persist correctly when navigating between pages.
**Validates: Requirements 6.1**

### Property 3: Navigation Integrity
*For any* category link clicked in the header, the system SHALL navigate to the correct collection page URL.
**Validates: Requirements 2.1, 4.5**

### Property 4: Payment Flow Continuity
*For any* checkout initiated from any collection page, the Midtrans payment flow SHALL complete successfully with correct order data.
**Validates: Requirements 6.2, 6.3, 6.4**

## Error Handling

1. **API Errors**: Display elegant error state with retry option
2. **Empty Category**: Show styled empty state with suggestions
3. **Image Load Failure**: Display placeholder with brand styling
4. **Payment Failure**: Redirect to order-failed page with clear messaging

## Testing Strategy

### Unit Tests
- Test category filtering logic
- Test cart operations across pages
- Test navigation routing

### Property-Based Tests
- Use fast-check library for TypeScript
- Minimum 100 iterations per property test
- Test category filter returns only matching products
- Test cart state consistency

### Integration Tests
- Test full checkout flow from category page
- Test Midtrans webhook handling
- Test database category queries

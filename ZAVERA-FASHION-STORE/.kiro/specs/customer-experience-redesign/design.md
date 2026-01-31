# Design Document

## Overview

This design enhances ZAVERA's customer experience to match premium fashion e-commerce standards. The focus is on adding product filtering/sorting, improving loading states, polishing the order history page, and ensuring consistent visual feedback throughout the shopping journey.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Enhanced Frontend                         │
├─────────────────────────────────────────────────────────────┤
│  Enhanced Components:                                        │
│  ├── CategoryPage (+ FilterPanel, SortDropdown)             │
│  ├── ProductCard (+ hover states, quick actions)            │
│  ├── ProductDetail (+ breadcrumb, image zoom, stock badge)  │
│  ├── CartPage (existing - already polished)                 │
│  ├── CheckoutPage (existing - already polished)             │
│  └── OrderHistoryPage (+ timeline, enhanced cards)          │
├─────────────────────────────────────────────────────────────┤
│  New Components:                                             │
│  ├── FilterPanel (size, price, subcategory filters)         │
│  ├── FilterDrawer (mobile slide-out version)                │
│  ├── ActiveFilters (removable filter tags)                  │
│  ├── OrderTimeline (visual status progression)              │
│  └── ImageZoom (hover zoom for product images)              │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Filter Panel Component
```typescript
interface FilterPanelProps {
  category: string;
  onFilterChange: (filters: ProductFilters) => void;
  activeFilters: ProductFilters;
}

interface ProductFilters {
  sizes: string[];
  priceRange: { min: number; max: number } | null;
  subcategory: string | null;
  sortBy: 'newest' | 'price-low' | 'price-high' | 'name';
}
```

### 2. Order Timeline Component
```typescript
interface OrderTimelineProps {
  status: OrderStatus;
  createdAt: string;
  updatedAt?: string;
}

type OrderStatus = 'PENDING' | 'PAID' | 'PROCESSING' | 'SHIPPED' | 'DELIVERED' | 'COMPLETED' | 'CANCELLED' | 'FAILED' | 'EXPIRED';

const statusSteps = [
  { key: 'PENDING', label: 'Menunggu Pembayaran', icon: 'clock' },
  { key: 'PAID', label: 'Dibayar', icon: 'check' },
  { key: 'PROCESSING', label: 'Diproses', icon: 'package' },
  { key: 'SHIPPED', label: 'Dikirim', icon: 'truck' },
  { key: 'DELIVERED', label: 'Terkirim', icon: 'home' },
];
```

### 3. Enhanced Category Page
```typescript
interface CategoryPageProps {
  category: ProductCategory;
  title: string;
  subtitle: string;
  bannerImage: string;
  subcategories: string[];
}
```

## Data Models

### Product Filters State
```typescript
interface FilterState {
  sizes: string[];
  minPrice: number | null;
  maxPrice: number | null;
  subcategory: string | null;
  sortBy: SortOption;
}

type SortOption = 'newest' | 'price-low' | 'price-high' | 'name';
```

### Order Display Model
```typescript
interface OrderDisplay {
  id: number;
  order_code: string;
  status: OrderStatus;
  statusLabel: string;
  statusColor: string;
  total_amount: number;
  created_at: string;
  items: OrderItemDisplay[];
  timeline: TimelineStep[];
}

interface TimelineStep {
  label: string;
  completed: boolean;
  current: boolean;
  date?: string;
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Filter Results Consistency
*For any* product list and any filter selection (size, price range, subcategory), all products in the filtered result SHALL match all active filter criteria.
**Validates: Requirements 1.2**

### Property 2: Sort Order Correctness
*For any* product list and sort option, the resulting array SHALL be correctly ordered (e.g., for price-low, each product's price <= next product's price).
**Validates: Requirements 1.3**

### Property 3: Cart Total Calculation
*For any* cart with items, the displayed total SHALL equal the sum of (price × quantity) for all items plus shipping cost.
**Validates: Requirements 4.2, 4.5**

### Property 4: Breadcrumb Category Match
*For any* product with a category, the breadcrumb navigation SHALL include that category in the hierarchy.
**Validates: Requirements 3.1**

### Property 5: Low Stock Indicator Display
*For any* product with stock less than the threshold (10), the low stock indicator SHALL be displayed.
**Validates: Requirements 3.4**

### Property 6: Order Card Information Completeness
*For any* order, the rendered order card SHALL contain the order code, formatted date, status badge, and formatted total amount.
**Validates: Requirements 6.1**

### Property 7: Order Timeline Status Progression
*For any* order status, the timeline SHALL show all steps up to and including the current status as completed.
**Validates: Requirements 6.2**

### Property 8: Pending Order Pay Button Visibility
*For any* order with PENDING status, the "Pay Now" button SHALL be visible and clickable.
**Validates: Requirements 6.4**

### Property 9: Status Badge Color Mapping
*For any* order status, the badge SHALL display the correct color and Indonesian label as defined in the status mapping.
**Validates: Requirements 6.5**

### Property 10: Form Validation Error Display
*For any* form field with invalid input, the field SHALL display an error message and visual error styling.
**Validates: Requirements 7.2**

### Property 11: Checkout Form Pre-fill
*For any* logged-in user entering checkout, the form fields SHALL be pre-populated with the user's stored information (name, email, phone).
**Validates: Requirements 5.1**

## Error Handling

1. **Filter No Results**: Display elegant empty state with suggestion to adjust filters
2. **API Errors**: Show error toast with retry option, maintain last known state
3. **Image Load Failure**: Display placeholder image with brand styling
4. **Order Load Failure**: Show error state with refresh button

## Testing Strategy

### Unit Tests
- Test filter logic for size, price, subcategory filtering
- Test sort functions for all sort options
- Test cart total calculation
- Test status badge color/label mapping
- Test timeline step calculation

### Property-Based Tests
Using fast-check library for TypeScript with minimum 100 iterations per test:

1. **Filter consistency test**: Generate random products and filters, verify all results match criteria
2. **Sort correctness test**: Generate random products, verify sort order is correct
3. **Cart total test**: Generate random cart items, verify total calculation
4. **Timeline progression test**: Generate random statuses, verify timeline steps
5. **Status badge test**: Generate all statuses, verify color/label mapping

### Integration Tests
- Test filter panel interaction with product grid
- Test order history page with various order states
- Test checkout form pre-fill for logged-in users

